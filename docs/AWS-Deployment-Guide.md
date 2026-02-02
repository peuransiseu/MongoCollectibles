# AWS Deployment Guide - MongoCollectibles Rental System

This guide provides three deployment architecture options for the MongoCollectibles rental system on AWS, ordered by cost efficiency, scalability, and ease of implementation.

---

## 📋 Table of Contents

1. [Architecture Options Overview](#architecture-options-overview)
2. [Option 1: Single EC2 Instance (Most Cost-Efficient)](#option-1-single-ec2-instance-most-cost-efficient)
3. [Option 2: Elastic Beanstalk (Easiest to Implement)](#option-2-elastic-beanstalk-easiest-to-implement)
4. [Option 3: ECS Fargate with RDS (Most Scalable)](#option-3-ecs-fargate-with-rds-most-scalable)
5. [Database Migration Strategy](#database-migration-strategy)
6. [Environment Configuration](#environment-configuration)
7. [Monitoring and Logging](#monitoring-and-logging)
8. [Cost Estimates](#cost-estimates)

---

## Architecture Options Overview

| Option | Cost/Month | Scalability | Ease of Setup | Best For |
|--------|-----------|-------------|---------------|----------|
| **EC2 Single Instance** | ~$10-20 | Low | Easy | Development, MVP, Low Traffic |
| **Elastic Beanstalk** | ~$20-40 | Medium | Easiest | Quick Production Deploy |
| **ECS Fargate + RDS** | ~$50-100 | High | Moderate | Production, High Traffic |

---

## Option 1: Single EC2 Instance (Most Cost-Efficient)

### Architecture Diagram

```mermaid
graph TB
    User[Users] --> Route53[Route 53<br/>DNS]
    Route53 --> ALB[Application Load Balancer<br/>Optional]
    ALB --> EC2[EC2 t3.micro Instance<br/>Go Application + Static Files]
    EC2 --> RDS[(RDS PostgreSQL<br/>t3.micro<br/>Optional)]
    
    style EC2 fill:#ff9900
    style RDS fill:#3b48cc
    style ALB fill:#8c4fff
```

### Components

- **EC2 Instance**: t3.micro (1 vCPU, 1GB RAM) - Free tier eligible
- **Elastic IP**: Static IP address
- **Security Group**: Controls inbound/outbound traffic
- **Optional**: Application Load Balancer for SSL/HTTPS
- **Optional**: RDS for persistent database

### Deployment Steps

#### 1. Prepare Your Application

Create a deployment script `deploy.sh`:

```bash
#!/bin/bash
# Build the Go application
GOOS=linux GOARCH=amd64 go build -o rental-system main.go

# Create deployment package
tar -czf deployment.tar.gz rental-system static/ .env
```

#### 2. Launch EC2 Instance

```bash
# Using AWS CLI
aws ec2 run-instances \
  --image-id ami-0c55b159cbfafe1f0 \
  --instance-type t3.micro \
  --key-name your-key-pair \
  --security-group-ids sg-xxxxxxxxx \
  --subnet-id subnet-xxxxxxxxx \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=MongoCollectibles}]'
```

Or use the AWS Console:
1. Go to EC2 Dashboard
2. Click "Launch Instance"
3. Select **Amazon Linux 2023** AMI
4. Choose **t3.micro** (free tier)
5. Configure security group:
   - SSH (22) - Your IP only
   - HTTP (80) - 0.0.0.0/0
   - HTTPS (443) - 0.0.0.0/0
   - Custom TCP (8080) - 0.0.0.0/0
6. Launch and download key pair

#### 3. Connect and Setup

```bash
# SSH into instance
ssh -i your-key.pem ec2-user@your-ec2-ip

# Update system
sudo yum update -y

# Install Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Create application directory
mkdir -p /home/ec2-user/mongocollectibles
cd /home/ec2-user/mongocollectibles
```

#### 4. Deploy Application

```bash
# From your local machine, copy files
scp -i your-key.pem deployment.tar.gz ec2-user@your-ec2-ip:/home/ec2-user/mongocollectibles/

# On EC2 instance
cd /home/ec2-user/mongocollectibles
tar -xzf deployment.tar.gz

# Create .env file
cat > .env << EOF
TEST_SECRET_KEY=your_paymongo_secret_key
TEST_PUBLIC_KEY=your_paymongo_public_key
SERVER_PORT=8080
ENVIRONMENT=production
EOF

# Make executable
chmod +x rental-system
```

#### 5. Create Systemd Service

```bash
sudo tee /etc/systemd/system/mongocollectibles.service > /dev/null << EOF
[Unit]
Description=MongoCollectibles Rental System
After=network.target

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/home/ec2-user/mongocollectibles
ExecStart=/home/ec2-user/mongocollectibles/rental-system
Restart=on-failure
RestartSec=5s
Environment="PATH=/usr/local/go/bin:/usr/bin:/bin"

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable mongocollectibles
sudo systemctl start mongocollectibles
sudo systemctl status mongocollectibles
```

#### 6. Setup Nginx as Reverse Proxy (Optional but Recommended)

```bash
# Install Nginx
sudo yum install nginx -y

# Configure Nginx
sudo tee /etc/nginx/conf.d/mongocollectibles.conf > /dev/null << EOF
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

# Start Nginx
sudo systemctl enable nginx
sudo systemctl start nginx
```

#### 7. Setup SSL with Let's Encrypt (Optional)

```bash
# Install Certbot
sudo yum install certbot python3-certbot-nginx -y

# Get SSL certificate
sudo certbot --nginx -d your-domain.com

# Auto-renewal is configured automatically
```

### Pros & Cons

**Pros:**
- ✅ Lowest cost (~$10-15/month)
- ✅ Simple architecture
- ✅ Full control over server
- ✅ Free tier eligible for 12 months

**Cons:**
- ❌ Manual scaling required
- ❌ Single point of failure
- ❌ Manual updates and maintenance
- ❌ No built-in load balancing

---

## Option 2: Elastic Beanstalk (Easiest to Implement)

### Architecture Diagram

```mermaid
graph TB
    User[Users] --> Route53[Route 53<br/>DNS]
    Route53 --> ALB[Application Load Balancer<br/>Auto-configured]
    ALB --> EB1[EC2 Instance 1<br/>Auto Scaling Group]
    ALB --> EB2[EC2 Instance 2<br/>Auto Scaling Group]
    EB1 --> RDS[(RDS PostgreSQL<br/>Managed)]
    EB2 --> RDS
    
    style EB1 fill:#ff9900
    style EB2 fill:#ff9900
    style RDS fill:#3b48cc
    style ALB fill:#8c4fff
```

### Components

- **Elastic Beanstalk Environment**: Manages EC2, Load Balancer, Auto Scaling
- **Application Load Balancer**: Automatically configured
- **Auto Scaling Group**: 1-4 instances based on load
- **RDS Database**: Optional, managed PostgreSQL
- **CloudWatch**: Automatic monitoring

### Deployment Steps

#### 1. Prepare Application for Elastic Beanstalk

Create `Buildfile` in project root:

```yaml
build:
  commands:
    - go build -o application main.go
```

Create `Procfile`:

```
web: ./application
```

Create `.ebextensions/01_environment.config`:

```yaml
option_settings:
  aws:elasticbeanstalk:application:environment:
    SERVER_PORT: 5000
    ENVIRONMENT: production
  aws:elasticbeanstalk:environment:proxy:
    ProxyServer: nginx
  aws:autoscaling:launchconfiguration:
    InstanceType: t3.micro
  aws:autoscaling:asg:
    MinSize: 1
    MaxSize: 4
  aws:elasticbeanstalk:environment:
    EnvironmentType: LoadBalanced
    LoadBalancerType: application
```

Create `.ebextensions/02_nginx.config`:

```yaml
files:
  /etc/nginx/conf.d/proxy.conf:
    mode: "000644"
    owner: root
    group: root
    content: |
      client_max_body_size 50M;
```

#### 2. Initialize Elastic Beanstalk

```bash
# Install EB CLI
pip install awsebcli

# Initialize EB application
eb init -p go mongocollectibles --region us-east-1

# Create environment
eb create mongocollectibles-prod \
  --instance-type t3.micro \
  --envvars TEST_SECRET_KEY=your_key,TEST_PUBLIC_KEY=your_key

# Deploy
eb deploy

# Open in browser
eb open
```

#### 3. Configure Environment Variables

```bash
# Set environment variables
eb setenv \
  TEST_SECRET_KEY=your_paymongo_secret_key \
  TEST_PUBLIC_KEY=your_paymongo_public_key \
  ENVIRONMENT=production

# Or use AWS Console:
# Elastic Beanstalk > Environment > Configuration > Software > Environment properties
```

#### 4. Setup Custom Domain

```bash
# Get load balancer URL
eb status

# In Route 53, create CNAME record:
# Name: www.your-domain.com
# Type: CNAME
# Value: your-eb-env.elasticbeanstalk.com
```

#### 5. Enable HTTPS

In AWS Console:
1. Go to EC2 > Load Balancers
2. Find your EB load balancer
3. Add listener on port 443
4. Request ACM certificate for your domain
5. Attach certificate to HTTPS listener

### Continuous Deployment with CodePipeline (Optional)

```yaml
# buildspec.yml
version: 0.2

phases:
  install:
    runtime-versions:
      golang: 1.21
  build:
    commands:
      - go build -o application main.go
artifacts:
  files:
    - application
    - Procfile
    - .ebextensions/**/*
    - static/**/*
```

### Pros & Cons

**Pros:**
- ✅ Easiest deployment (one command)
- ✅ Auto-scaling built-in
- ✅ Load balancer included
- ✅ Automatic health monitoring
- ✅ Easy rollback
- ✅ Managed platform updates

**Cons:**
- ❌ Higher cost than single EC2
- ❌ Less control over infrastructure
- ❌ Can be overkill for small apps

---

## Option 3: ECS Fargate with RDS (Most Scalable)

### Architecture Diagram

```mermaid
graph TB
    User[Users] --> Route53[Route 53<br/>DNS]
    Route53 --> ALB[Application Load Balancer]
    ALB --> ECS1[ECS Fargate Task 1<br/>Container]
    ALB --> ECS2[ECS Fargate Task 2<br/>Container]
    ECS1 --> RDS[(RDS PostgreSQL<br/>Multi-AZ)]
    ECS2 --> RDS
    ECS1 --> S3[S3 Bucket<br/>Static Assets]
    ECS2 --> S3
    
    CloudWatch[CloudWatch Logs] -.-> ECS1
    CloudWatch -.-> ECS2
    
    style ECS1 fill:#ff9900
    style ECS2 fill:#ff9900
    style RDS fill:#3b48cc
    style ALB fill:#8c4fff
    style S3 fill:#569a31
```

### Components

- **ECS Fargate**: Serverless container orchestration
- **Application Load Balancer**: Traffic distribution
- **RDS PostgreSQL**: Managed database with Multi-AZ
- **S3**: Static file hosting (optional)
- **CloudWatch**: Logs and monitoring
- **ECR**: Container registry
- **Auto Scaling**: Based on CPU/memory metrics

### Deployment Steps

#### 1. Create Dockerfile

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rental-system .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary and static files
COPY --from=builder /app/rental-system .
COPY --from=builder /app/static ./static

# Expose port
EXPOSE 8080

# Run application
CMD ["./rental-system"]
```

Create `.dockerignore`:

```
.git
.env
*.md
bin/
*.exe
deployment.tar.gz
```

#### 2. Build and Push to ECR

```bash
# Authenticate Docker to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

# Create ECR repository
aws ecr create-repository --repository-name mongocollectibles

# Build Docker image
docker build -t mongocollectibles .

# Tag image
docker tag mongocollectibles:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/mongocollectibles:latest

# Push to ECR
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/mongocollectibles:latest
```

#### 3. Create RDS Database

```bash
# Create DB subnet group
aws rds create-db-subnet-group \
  --db-subnet-group-name mongocollectibles-subnet-group \
  --db-subnet-group-description "Subnet group for MongoCollectibles" \
  --subnet-ids subnet-xxxxx subnet-yyyyy

# Create RDS instance
aws rds create-db-instance \
  --db-instance-identifier mongocollectibles-db \
  --db-instance-class db.t3.micro \
  --engine postgres \
  --engine-version 15.4 \
  --master-username admin \
  --master-user-password YourSecurePassword123! \
  --allocated-storage 20 \
  --db-subnet-group-name mongocollectibles-subnet-group \
  --vpc-security-group-ids sg-xxxxxxxxx \
  --backup-retention-period 7 \
  --multi-az
```

#### 4. Create ECS Task Definition

Create `task-definition.json`:

```json
{
  "family": "mongocollectibles",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "containerDefinitions": [
    {
      "name": "mongocollectibles-app",
      "image": "<account-id>.dkr.ecr.us-east-1.amazonaws.com/mongocollectibles:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "SERVER_PORT",
          "value": "8080"
        },
        {
          "name": "ENVIRONMENT",
          "value": "production"
        }
      ],
      "secrets": [
        {
          "name": "TEST_SECRET_KEY",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:xxxxx:secret:paymongo-secret"
        },
        {
          "name": "TEST_PUBLIC_KEY",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:xxxxx:secret:paymongo-public"
        },
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:xxxxx:secret:db-connection"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/mongocollectibles",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

Register task definition:

```bash
aws ecs register-task-definition --cli-input-json file://task-definition.json
```

#### 5. Create ECS Cluster and Service

```bash
# Create cluster
aws ecs create-cluster --cluster-name mongocollectibles-cluster

# Create service
aws ecs create-service \
  --cluster mongocollectibles-cluster \
  --service-name mongocollectibles-service \
  --task-definition mongocollectibles \
  --desired-count 2 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxxxx,subnet-yyyyy],securityGroups=[sg-xxxxxxxxx],assignPublicIp=ENABLED}" \
  --load-balancers "targetGroupArn=arn:aws:elasticloadbalancing:...,containerName=mongocollectibles-app,containerPort=8080"
```

#### 6. Setup Auto Scaling

```bash
# Register scalable target
aws application-autoscaling register-scalable-target \
  --service-namespace ecs \
  --scalable-dimension ecs:service:DesiredCount \
  --resource-id service/mongocollectibles-cluster/mongocollectibles-service \
  --min-capacity 2 \
  --max-capacity 10

# Create scaling policy
aws application-autoscaling put-scaling-policy \
  --service-namespace ecs \
  --scalable-dimension ecs:service:DesiredCount \
  --resource-id service/mongocollectibles-cluster/mongocollectibles-service \
  --policy-name cpu-scaling-policy \
  --policy-type TargetTrackingScaling \
  --target-tracking-scaling-policy-configuration file://scaling-policy.json
```

`scaling-policy.json`:

```json
{
  "TargetValue": 70.0,
  "PredefinedMetricSpecification": {
    "PredefinedMetricType": "ECSServiceAverageCPUUtilization"
  },
  "ScaleInCooldown": 300,
  "ScaleOutCooldown": 60
}
```

### Pros & Cons

**Pros:**
- ✅ Highly scalable
- ✅ No server management
- ✅ Pay only for what you use
- ✅ Easy container updates
- ✅ Built-in service discovery
- ✅ Production-ready architecture

**Cons:**
- ❌ Higher cost
- ❌ More complex setup
- ❌ Requires Docker knowledge
- ❌ Cold start times (minimal)

---

## Database Migration Strategy

Currently, the application uses in-memory data storage. For production, migrate to a persistent database.

### Recommended: PostgreSQL on RDS

#### 1. Update Go Dependencies

```bash
go get github.com/lib/pq
go get gorm.io/gorm
go get gorm.io/driver/postgres
```

#### 2. Create Database Schema

```sql
-- Create tables
CREATE TABLE collectibles (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    size VARCHAR(20),
    image_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE warehouses (
    id VARCHAR(50) PRIMARY KEY,
    distances_to_stores JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE collectible_units (
    id VARCHAR(50) PRIMARY KEY,
    collectible_id VARCHAR(50) REFERENCES collectibles(id),
    warehouse_id VARCHAR(50) REFERENCES warehouses(id),
    is_available BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE rentals (
    id VARCHAR(50) PRIMARY KEY,
    collectible_id VARCHAR(50) REFERENCES collectibles(id),
    collectible_name VARCHAR(255),
    store_id VARCHAR(50),
    warehouse_id VARCHAR(50),
    customer JSONB,
    duration INTEGER,
    daily_rate DECIMAL(10,2),
    total_fee DECIMAL(10,2),
    payment_method VARCHAR(50),
    payment_status VARCHAR(20),
    payment_id VARCHAR(100),
    payment_url TEXT,
    eta INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rentals_payment_id ON rentals(payment_id);
CREATE INDEX idx_collectible_units_available ON collectible_units(is_available);
```

#### 3. Update Repository Layer

Modify `data/repository.go` to use GORM instead of in-memory storage:

```go
package data

import (
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
    "github.com/mongocollectibles/rental-system/models"
)

type Repository struct {
    db *gorm.DB
}

func NewRepository(databaseURL string) (*Repository, error) {
    db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    
    // Auto-migrate tables
    db.AutoMigrate(&models.Collectible{}, &models.Rental{}, &models.CollectibleUnit{})
    
    return &Repository{db: db}, nil
}

func (r *Repository) GetAllCollectibles() []*models.Collectible {
    var collectibles []*models.Collectible
    r.db.Find(&collectibles)
    return collectibles
}

// ... implement other methods
```

#### 4. Update main.go

```go
func main() {
    cfg := config.LoadConfig()
    
    // Connect to database
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        databaseURL = "host=localhost user=postgres password=postgres dbname=mongocollectibles port=5432 sslmode=disable"
    }
    
    repo, err := data.NewRepository(databaseURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    // Seed data only if database is empty
    if repo.IsEmpty() {
        data.SeedData(repo)
    }
    
    // ... rest of initialization
}
```

---

## Environment Configuration

### Using AWS Secrets Manager (Recommended for Production)

```bash
# Store PayMongo secrets
aws secretsmanager create-secret \
  --name mongocollectibles/paymongo \
  --secret-string '{"secret_key":"sk_test_xxx","public_key":"pk_test_xxx"}'

# Store database credentials
aws secretsmanager create-secret \
  --name mongocollectibles/database \
  --secret-string '{"url":"postgresql://user:pass@host:5432/dbname"}'
```

### Environment Variables by Deployment Option

| Variable | EC2 | Elastic Beanstalk | ECS Fargate |
|----------|-----|-------------------|-------------|
| `TEST_SECRET_KEY` | .env file | EB Environment | Secrets Manager |
| `TEST_PUBLIC_KEY` | .env file | EB Environment | Secrets Manager |
| `DATABASE_URL` | .env file | EB Environment | Secrets Manager |
| `SERVER_PORT` | .env file | EB Config | Task Definition |
| `ENVIRONMENT` | .env file | EB Environment | Task Definition |

---

## Monitoring and Logging

### CloudWatch Setup

#### 1. Create Log Group

```bash
aws logs create-log-group --log-group-name /aws/mongocollectibles
```

#### 2. Create CloudWatch Dashboard

```bash
aws cloudwatch put-dashboard \
  --dashboard-name MongoCollectibles \
  --dashboard-body file://dashboard.json
```

`dashboard.json`:

```json
{
  "widgets": [
    {
      "type": "metric",
      "properties": {
        "metrics": [
          ["AWS/ApplicationELB", "TargetResponseTime", {"stat": "Average"}],
          [".", "RequestCount", {"stat": "Sum"}],
          [".", "HTTPCode_Target_4XX_Count", {"stat": "Sum"}],
          [".", "HTTPCode_Target_5XX_Count", {"stat": "Sum"}]
        ],
        "period": 300,
        "stat": "Average",
        "region": "us-east-1",
        "title": "Application Performance"
      }
    }
  ]
}
```

#### 3. Setup Alarms

```bash
# High error rate alarm
aws cloudwatch put-metric-alarm \
  --alarm-name mongocollectibles-high-errors \
  --alarm-description "Alert when error rate is high" \
  --metric-name HTTPCode_Target_5XX_Count \
  --namespace AWS/ApplicationELB \
  --statistic Sum \
  --period 300 \
  --threshold 10 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2

# High CPU alarm
aws cloudwatch put-metric-alarm \
  --alarm-name mongocollectibles-high-cpu \
  --alarm-description "Alert when CPU is high" \
  --metric-name CPUUtilization \
  --namespace AWS/ECS \
  --statistic Average \
  --period 300 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2
```

### Application Logging

Update your Go application to use structured logging:

```go
import (
    "log"
    "os"
)

func main() {
    // Configure JSON logging for CloudWatch
    log.SetOutput(os.Stdout)
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    
    // Log application start
    log.Println("Starting MongoCollectibles Rental System")
    
    // ... rest of code
}
```

---

## Cost Estimates

### Monthly Cost Breakdown

#### Option 1: Single EC2 Instance

| Service | Specification | Monthly Cost |
|---------|--------------|--------------|
| EC2 t3.micro | 1 vCPU, 1GB RAM | $7.50 |
| Elastic IP | 1 static IP | $3.60 |
| EBS Storage | 20GB gp3 | $1.60 |
| Data Transfer | 10GB outbound | $0.90 |
| **Total** | | **~$13.60** |

*First year free tier: $0*

#### Option 2: Elastic Beanstalk

| Service | Specification | Monthly Cost |
|---------|--------------|--------------|
| EC2 Instances | 2x t3.micro | $15.00 |
| Application Load Balancer | 1 ALB | $16.20 |
| EBS Storage | 2x 20GB | $3.20 |
| Data Transfer | 50GB outbound | $4.50 |
| RDS t3.micro (optional) | PostgreSQL | $15.00 |
| **Total** | | **~$38.90 - $53.90** |

#### Option 3: ECS Fargate + RDS

| Service | Specification | Monthly Cost |
|---------|--------------|--------------|
| Fargate Tasks | 2x 0.25 vCPU, 0.5GB | $14.40 |
| Application Load Balancer | 1 ALB | $16.20 |
| RDS db.t3.micro | PostgreSQL Multi-AZ | $30.00 |
| ECR Storage | 1GB | $0.10 |
| CloudWatch Logs | 5GB | $2.50 |
| Data Transfer | 50GB outbound | $4.50 |
| **Total** | | **~$67.70** |

### Cost Optimization Tips

1. **Use Reserved Instances**: Save up to 72% for EC2/RDS with 1-3 year commitments
2. **Enable Auto Scaling**: Scale down during off-peak hours
3. **Use S3 for Static Assets**: Cheaper than serving from EC2
4. **Implement CloudFront CDN**: Reduce data transfer costs
5. **Right-size Instances**: Monitor and adjust instance types based on actual usage
6. **Use Spot Instances**: For non-critical workloads (up to 90% savings)

---

## Recommendation

### For MVP/Development
**Choose Option 1 (Single EC2)** - Cost-effective, simple, and sufficient for initial launch.

### For Quick Production Deploy
**Choose Option 2 (Elastic Beanstalk)** - Easiest to set up with built-in scaling and monitoring.

### For High-Traffic Production
**Choose Option 3 (ECS Fargate)** - Best scalability, reliability, and modern architecture.

---

## Next Steps

1. **Choose your architecture** based on requirements and budget
2. **Set up AWS account** and configure billing alerts
3. **Migrate to persistent database** (PostgreSQL on RDS)
4. **Update PayMongo URLs** to use production domain
5. **Implement CI/CD pipeline** for automated deployments
6. **Set up monitoring and alerts**
7. **Configure backups** for database and application
8. **Perform load testing** before going live

---

## Additional Resources

- [AWS Well-Architected Framework](https://aws.amazon.com/architecture/well-architected/)
- [AWS Cost Calculator](https://calculator.aws/)
- [Go on AWS Best Practices](https://aws.amazon.com/blogs/compute/deploying-go-applications-to-aws/)
- [PayMongo Documentation](https://developers.paymongo.com/)

---

**Need Help?** Consider using AWS Support or consulting with an AWS Solutions Architect for production deployments.
