# AWS Deployment Guide for MongoCollectibles

## Overview
This guide provides multiple AWS deployment architectures for the MongoCollectibles rental system, ranging from cost-efficient solutions for startups to scalable enterprise-grade deployments.

---

## Architecture Options

### 🌟 Option 1: Minimal Cost Architecture (Recommended for MVP)

**Estimated Monthly Cost**: $15-30 USD

#### Architecture Diagram
```
┌─────────────────┐
│   CloudFront    │ ← CDN for static assets
│   (Free Tier)   │
└────────┬────────┘
         │
┌────────▼────────┐
│   EC2 t3.micro  │ ← Go application + static files
│   ($7-10/month) │ ← Free tier eligible (12 months)
└────────┬────────┘
         │
┌────────▼────────┐
│  DynamoDB       │ ← NoSQL database
│  (Free Tier)    │ ← 25GB storage, 25 WCU/RCU
└─────────────────┘
```

#### Components

**1. EC2 t3.micro Instance**
- **Purpose**: Run Go application
- **Specs**: 2 vCPU, 1GB RAM
- **Cost**: Free tier (12 months), then ~$7-10/month
- **Setup**:
  ```bash
  # Install Go
  sudo yum install golang -y
  
  # Clone repository
  git clone <your-repo>
  cd MongoCollectibles
  
  # Build application
  go build -o mongocollectibles main.go
  
  # Run with systemd
  sudo systemctl enable mongocollectibles
  sudo systemctl start mongocollectibles
  ```

**2. DynamoDB**
- **Purpose**: Store collectibles, rentals, warehouses
- **Cost**: Free tier (25GB, 25 WCU/RCU)
- **Tables**:
  - `Collectibles` (PK: id)
  - `Rentals` (PK: id, GSI: customer_email)
  - `Warehouses` (PK: collectible_id, SK: warehouse_id)

**3. CloudFront**
- **Purpose**: CDN for static assets (images, CSS, JS)
- **Cost**: Free tier (1TB transfer/month)

**4. Route 53**
- **Purpose**: DNS management
- **Cost**: $0.50/month per hosted zone

**5. Certificate Manager**
- **Purpose**: Free SSL/TLS certificates
- **Cost**: Free

#### Pros
✅ Lowest cost option  
✅ Free tier eligible  
✅ Simple to manage  
✅ Good for MVP/testing  

#### Cons
❌ Single point of failure  
❌ Manual scaling required  
❌ Limited to 1 instance  

---

### 💼 Option 2: Production-Ready Architecture

**Estimated Monthly Cost**: $80-150 USD

#### Architecture Diagram
```
                    ┌─────────────┐
                    │  Route 53   │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ CloudFront  │
                    └──────┬──────┘
                           │
                    ┌──────▼──────────────┐
                    │  Application LB     │
                    │  ($16/month)        │
                    └──────┬──────────────┘
                           │
              ┌────────────┴────────────┐
              │                         │
       ┌──────▼──────┐          ┌──────▼──────┐
       │ EC2 t3.small│          │ EC2 t3.small│
       │ (AZ-1)      │          │ (AZ-2)      │
       │ $15/month   │          │ $15/month   │
       └──────┬──────┘          └──────┬──────┘
              │                         │
              └────────────┬────────────┘
                           │
                    ┌──────▼──────────┐
                    │   RDS Aurora    │
                    │   Serverless v2 │
                    │   ($30-60/mo)   │
                    └─────────────────┘
```

#### Components

**1. Application Load Balancer (ALB)**
- **Purpose**: Distribute traffic, health checks
- **Cost**: ~$16/month
- **Features**:
  - SSL termination
  - Path-based routing
  - Health checks

**2. EC2 Auto Scaling Group**
- **Instances**: 2× t3.small (min), 4× t3.small (max)
- **Cost**: $15/instance/month
- **Configuration**:
  ```yaml
  MinSize: 2
  MaxSize: 4
  DesiredCapacity: 2
  HealthCheckType: ELB
  HealthCheckGracePeriod: 300
  ```

**3. RDS Aurora Serverless v2**
- **Purpose**: PostgreSQL-compatible database
- **Cost**: $30-60/month (0.5-1 ACU)
- **Features**:
  - Auto-scaling
  - Multi-AZ replication
  - Automated backups

**4. S3 + CloudFront**
- **Purpose**: Static asset hosting
- **Cost**: $5-10/month
- **Setup**:
  ```bash
  # Sync static files to S3
  aws s3 sync ./static s3://mongocollectibles-static/
  
  # Invalidate CloudFront cache
  aws cloudfront create-invalidation --distribution-id XXX --paths "/*"
  ```

**5. ElastiCache Redis**
- **Purpose**: Session storage, caching
- **Cost**: $15/month (t3.micro)
- **Use Cases**:
  - API response caching
  - Rate limiting
  - Session management

#### Pros
✅ High availability (Multi-AZ)  
✅ Auto-scaling  
✅ Managed database  
✅ Production-ready  

#### Cons
❌ Higher cost  
❌ More complex setup  

---

### 🚀 Option 3: Serverless Architecture

**Estimated Monthly Cost**: $20-50 USD (pay-per-use)

#### Architecture Diagram
```
┌─────────────┐
│  Route 53   │
└──────┬──────┘
       │
┌──────▼──────────┐
│  API Gateway    │ ← REST API
│  ($3.50/million)│
└──────┬──────────┘
       │
┌──────▼──────────┐
│  Lambda (Go)    │ ← Serverless functions
│  ($0.20/million)│
└──────┬──────────┘
       │
┌──────▼──────────┐
│  DynamoDB       │ ← NoSQL database
│  (On-demand)    │
└─────────────────┘
       │
┌──────▼──────────┐
│  S3 + CloudFront│ ← Static hosting
│  ($5/month)     │
└─────────────────┘
```

#### Components

**1. API Gateway**
- **Purpose**: REST API endpoint
- **Cost**: $3.50 per million requests
- **Routes**:
  - `GET /collectibles`
  - `POST /rentals/quote`
  - `POST /rentals/checkout`

**2. Lambda Functions**
- **Runtime**: Go 1.x
- **Cost**: $0.20 per million requests
- **Functions**:
  ```
  - GetCollectibles (128MB, 3s timeout)
  - GetCollectibleById (128MB, 3s timeout)
  - CalculateQuote (256MB, 5s timeout)
  - ProcessCheckout (512MB, 10s timeout)
  - PayMongoWebhook (256MB, 10s timeout)
  ```

**3. DynamoDB (On-Demand)**
- **Cost**: $1.25 per million write requests
- **Advantages**:
  - No capacity planning
  - Auto-scaling
  - Pay only for what you use

**4. S3 + CloudFront**
- **Purpose**: Host static frontend
- **Cost**: ~$5/month
- **Configuration**:
  ```json
  {
    "IndexDocument": "index.html",
    "ErrorDocument": "index.html",
    "RoutingRules": []
  }
  ```

**5. Secrets Manager**
- **Purpose**: Store PayMongo API keys
- **Cost**: $0.40 per secret/month

#### Migration Steps

**1. Refactor to Lambda Handlers**
```go
// lambda/handlers/collectibles.go
package main

import (
    "context"
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)

func HandleGetCollectibles(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    // Your existing handler logic
    return events.APIGatewayProxyResponse{
        StatusCode: 200,
        Body: `{"success": true, "data": [...]}`,
        Headers: map[string]string{
            "Content-Type": "application/json",
        },
    }, nil
}

func main() {
    lambda.Start(HandleGetCollectibles)
}
```

**2. Deploy with SAM**
```yaml
# template.yaml
AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31

Resources:
  GetCollectiblesFunction:
    Type: AWS::Serverless::Function
    Properties:
      CodeUri: lambda/collectibles/
      Handler: bootstrap
      Runtime: provided.al2
      Architectures:
        - arm64
      Events:
        GetCollectibles:
          Type: Api
          Properties:
            Path: /api/collectibles
            Method: get
```

#### Pros
✅ Lowest operational overhead  
✅ Auto-scaling built-in  
✅ Pay only for usage  
✅ No server management  

#### Cons
❌ Cold start latency  
❌ Requires code refactoring  
❌ 15-minute Lambda timeout limit  

---

### 🏢 Option 4: Container-Based (ECS Fargate)

**Estimated Monthly Cost**: $50-100 USD

#### Architecture Diagram
```
┌─────────────┐
│  Route 53   │
└──────┬──────┘
       │
┌──────▼──────────┐
│  Application LB │
└──────┬──────────┘
       │
┌──────▼──────────┐
│  ECS Fargate    │ ← Containers (0.25 vCPU, 0.5GB)
│  2 tasks        │ ← $0.04/hour per task
│  ($30/month)    │
└──────┬──────────┘
       │
┌──────▼──────────┐
│  Aurora         │
│  Serverless v2  │
└─────────────────┘
```

#### Components

**1. Dockerfile**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o mongocollectibles main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mongocollectibles .
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["./mongocollectibles"]
```

**2. ECS Task Definition**
```json
{
  "family": "mongocollectibles",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "containerDefinitions": [{
    "name": "app",
    "image": "123456789.dkr.ecr.us-east-1.amazonaws.com/mongocollectibles:latest",
    "portMappings": [{
      "containerPort": 8080,
      "protocol": "tcp"
    }],
    "environment": [
      {"name": "SERVER_PORT", "value": "8080"}
    ],
    "secrets": [
      {"name": "PAYMONGO_SECRET_KEY", "valueFrom": "arn:aws:secretsmanager:..."}
    ]
  }]
}
```

**3. Deployment**
```bash
# Build and push image
docker build -t mongocollectibles .
aws ecr get-login-password | docker login --username AWS --password-stdin 123456789.dkr.ecr.us-east-1.amazonaws.com
docker tag mongocollectibles:latest 123456789.dkr.ecr.us-east-1.amazonaws.com/mongocollectibles:latest
docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/mongocollectibles:latest

# Update ECS service
aws ecs update-service --cluster mongocollectibles --service app --force-new-deployment
```

#### Pros
✅ Easy container deployment  
✅ No server management  
✅ Good for microservices  
✅ CI/CD friendly  

#### Cons
❌ More expensive than EC2  
❌ Requires Docker knowledge  

---

## Database Migration Strategies

### Option A: DynamoDB (NoSQL)

**Schema Design**:
```javascript
// Collectibles Table
{
  "id": "col-001",  // Partition Key
  "name": "Batman Figure",
  "size": "S",
  "image_url": "/images/batman.jpg",
  "warehouses": [
    {"id": "wh-001-1", "distances": [1,4,5]},
    {"id": "wh-001-2", "distances": [3,2,3]}
  ]
}

// Rentals Table
{
  "id": "rental-123",  // Partition Key
  "customer_email": "test@example.com",  // GSI
  "collectible_id": "col-001",
  "store_id": "store-a",
  "created_at": "2026-01-27T00:00:00Z",
  "payment_status": "completed"
}
```

**Go SDK Integration**:
```go
import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/dynamodb"
)

func (r *Repository) GetCollectibleByID(id string) (*models.Collectible, error) {
    result, err := r.dynamoClient.GetItem(&dynamodb.GetItemInput{
        TableName: aws.String("Collectibles"),
        Key: map[string]*dynamodb.AttributeValue{
            "id": {S: aws.String(id)},
        },
    })
    // Parse result...
}
```

---

### Option B: RDS Aurora PostgreSQL

**Schema**:
```sql
CREATE TABLE collectibles (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    size VARCHAR(1) CHECK (size IN ('S', 'M', 'L')),
    image_url VARCHAR(255),
    available BOOLEAN DEFAULT true
);

CREATE TABLE warehouses (
    id VARCHAR(50) PRIMARY KEY,
    collectible_id VARCHAR(50) REFERENCES collectibles(id),
    name VARCHAR(255),
    available BOOLEAN DEFAULT true,
    distances_to_stores INTEGER[]
);

CREATE TABLE rentals (
    id VARCHAR(50) PRIMARY KEY,
    collectible_id VARCHAR(50) REFERENCES collectibles(id),
    store_id VARCHAR(50),
    customer_email VARCHAR(255),
    duration INTEGER,
    total_fee DECIMAL(10,2),
    payment_status VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_rentals_customer ON rentals(customer_email);
CREATE INDEX idx_rentals_status ON rentals(payment_status);
```

**Go Integration**:
```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

func (r *Repository) GetAllCollectibles() ([]*models.Collectible, error) {
    rows, err := r.db.Query("SELECT id, name, size, image_url FROM collectibles WHERE available = true")
    // Parse rows...
}
```

---

## Alternative Cloud Providers

### Google Cloud Platform (GCP)

**Architecture**:
```
Cloud Run (Serverless containers) → $0.00002400/vCPU-second
Cloud SQL (PostgreSQL) → $7/month (db-f1-micro)
Cloud CDN → $0.08/GB
Cloud Storage → $0.020/GB
```

**Estimated Cost**: $30-60/month

**Advantages**:
- Generous free tier
- Excellent container support
- Strong AI/ML integration

---

### Microsoft Azure

**Architecture**:
```
App Service (B1) → $13/month
Azure Database for PostgreSQL → $30/month
Azure CDN → $0.081/GB
Blob Storage → $0.018/GB
```

**Estimated Cost**: $50-80/month

**Advantages**:
- Good Windows integration
- Enterprise features
- Hybrid cloud options

---

### DigitalOcean (Budget-Friendly)

**Architecture**:
```
Droplet (2GB RAM) → $12/month
Managed PostgreSQL → $15/month
Spaces (CDN + Storage) → $5/month
```

**Estimated Cost**: $32/month

**Advantages**:
- Simple pricing
- Easy to use
- Great for startups
- Predictable costs

---

### Vercel + Supabase (Modern Stack)

**Architecture**:
```
Vercel (Frontend + API) → Free tier / $20/month
Supabase (PostgreSQL + Auth) → Free tier / $25/month
```

**Estimated Cost**: $0-45/month

**Setup**:
```bash
# Deploy frontend to Vercel
vercel deploy

# Use Supabase for database
# API routes in /api directory
```

**Advantages**:
- Excellent DX
- Built-in auth
- Real-time features
- Generous free tier

---

## Cost Comparison Summary

| Solution | Monthly Cost | Best For | Scalability |
|----------|--------------|----------|-------------|
| **AWS EC2 + DynamoDB** | $15-30 | MVP, Testing | Manual |
| **AWS Production (ALB+EC2+RDS)** | $80-150 | Production | Auto-scale |
| **AWS Serverless** | $20-50 | Variable traffic | Infinite |
| **AWS ECS Fargate** | $50-100 | Containers | Auto-scale |
| **GCP Cloud Run** | $30-60 | Serverless containers | Auto-scale |
| **Azure App Service** | $50-80 | Enterprise | Auto-scale |
| **DigitalOcean** | $32 | Startups | Manual |
| **Vercel + Supabase** | $0-45 | Modern apps | Auto-scale |

---

## Recommended Deployment Path

### Phase 1: MVP (Month 1-3)
**Platform**: AWS EC2 t3.micro + DynamoDB  
**Cost**: $15-30/month  
**Why**: Free tier eligible, simple setup, low cost

### Phase 2: Growth (Month 4-12)
**Platform**: AWS ALB + Auto Scaling + RDS Aurora  
**Cost**: $80-150/month  
**Why**: High availability, auto-scaling, production-ready

### Phase 3: Scale (Year 2+)
**Platform**: AWS ECS Fargate + Aurora + CloudFront  
**Cost**: $200-500/month  
**Why**: Container orchestration, global CDN, enterprise features

---

## Additional AWS Services to Consider

### 1. **AWS WAF** (Web Application Firewall)
- **Cost**: $5/month + $1/million requests
- **Purpose**: DDoS protection, bot mitigation

### 2. **Amazon SES** (Simple Email Service)
- **Cost**: $0.10 per 1,000 emails
- **Purpose**: Rental confirmation emails

### 3. **Amazon SNS** (Simple Notification Service)
- **Cost**: $0.50 per million requests
- **Purpose**: Payment webhooks, alerts

### 4. **AWS CloudWatch**
- **Cost**: Free tier, then $0.30/metric/month
- **Purpose**: Monitoring, logging, alarms

### 5. **AWS Backup**
- **Cost**: $0.05/GB/month
- **Purpose**: Automated database backups

---

## Security Best Practices

### 1. **VPC Configuration**
```
Public Subnet (ALB only)
Private Subnet (EC2, RDS)
NAT Gateway for outbound traffic
```

### 2. **IAM Roles**
```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "secretsmanager:GetSecretValue"
    ],
    "Resource": "*"
  }]
}
```

### 3. **Secrets Management**
- Store PayMongo keys in AWS Secrets Manager
- Rotate secrets every 90 days
- Use IAM roles, not access keys

### 4. **SSL/TLS**
- Use AWS Certificate Manager (free)
- Enforce HTTPS only
- TLS 1.2+ minimum

---

## CI/CD Pipeline

### GitHub Actions + AWS
```yaml
name: Deploy to AWS
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v1
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1
      
      - name: Build Go app
        run: |
          go build -o mongocollectibles main.go
      
      - name: Deploy to EC2
        run: |
          scp -i key.pem mongocollectibles ec2-user@${{ secrets.EC2_HOST }}:/app/
          ssh -i key.pem ec2-user@${{ secrets.EC2_HOST }} 'sudo systemctl restart mongocollectibles'
```

---

## Conclusion

**For MongoCollectibles, I recommend**:

1. **Start with**: AWS EC2 t3.micro + DynamoDB (Free tier)
2. **Migrate to**: AWS ALB + Auto Scaling + RDS Aurora (when traffic grows)
3. **Consider**: Serverless (Lambda) if traffic is highly variable

This approach minimizes initial costs while providing a clear path to scale as your business grows.
