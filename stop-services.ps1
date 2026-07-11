# Stop all Money Planner services

Write-Host "Stopping Money Planner services..." -ForegroundColor Yellow

# Kill backend process (statement-import-api)
Write-Host "Stopping backend..." -ForegroundColor Cyan
Get-Process -Name "statement-import-api" -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Seconds 1

# Kill frontend process (node)
Write-Host "Stopping frontend..." -ForegroundColor Cyan
Get-Process -Name "node" -ErrorAction SilentlyContinue | Where-Object { $_.CommandLine -match "next|frontend" } | Stop-Process -Force
Start-Sleep -Seconds 1

# Stop docker-compose services
Write-Host "Stopping Docker Compose services..." -ForegroundColor Cyan
$projectDir = "C:\Users\kadaro\OneDrive - Autodesk\Desktop\repos\money-planner"
Set-Location $projectDir
docker-compose down -v

Write-Host "All services stopped!" -ForegroundColor Green
Write-Host ""
Write-Host "To restart:" -ForegroundColor Yellow
Write-Host "  1. docker-compose up -d"
Write-Host "  2. cd backend && go run ./cmd/statement-import-api"
Write-Host "  3. cd frontend && npm run dev"
