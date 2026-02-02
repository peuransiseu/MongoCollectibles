# deploy_app.ps1
param (
    [Parameter(Mandatory = $true)]
    [string]$BucketName,
    
    [Parameter(Mandatory = $false)]
    [string]$Profile = "default"
)

# 0. Clear conflicting environment variables that might override the profile
$env:AWS_ACCESS_KEY_ID = $null
$env:AWS_SECRET_ACCESS_KEY = $null
$env:AWS_SESSION_TOKEN = $null

# 1. Build for Linux
Write-Host "Building Go application for Linux..."
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -o rental-system main.go
$env:GOOS = $null
$env:CGO_ENABLED = $null

if (-not (Test-Path "rental-system")) {
    Write-Error "Build failed!"
    exit 1
}

# 2. Open AWS connection (verify connectivity)
Write-Host "Verifying AWS access using profile: $Profile..."
aws sts get-caller-identity --profile $Profile | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Error "AWS CLI not configured or not logged in for profile '$Profile'."
    exit 1
}

# 3. Create Bucket (idempotent-ish)
Write-Host "Checking S3 Bucket: $BucketName"
if (aws s3 ls "s3://$BucketName" --profile $Profile 2>&1 | Select-String -Pattern "NoSuchBucket") {
    Write-Host "Creating bucket $BucketName..."
    aws s3 mb "s3://$BucketName" --profile $Profile
}

# 4. Upload Files
Write-Host "Uploading binary..."
aws s3 cp rental-system "s3://$BucketName/rental-system" --profile $Profile

Write-Host "Uploading static files..."
aws s3 sync static "s3://$BucketName/static" --profile $Profile

Write-Host "Deployment artifacts uploaded successfully!"
Write-Host "Deploying CloudFormation Stack (This may take a few minutes)..."
aws cloudformation deploy --template-file template.yaml --stack-name MongoCollectibles --capabilities CAPABILITY_IAM --parameter-overrides KeyName=mongoapp ArtifactBucket=$BucketName --profile $Profile

if ($LASTEXITCODE -eq 0) {
    Write-Host "Stack update initiated successfully."
}
else {
    Write-Warning "Stack update finished with status $LASTEXITCODE (Exit code 255 usually means 'No changes to deploy', which is fine)."
}
