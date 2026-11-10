$health = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get
Write-Output "Health: $($health | ConvertTo-Json -Compress)"

$loginBody = '{"email":"it.admin@company.com","password":"Admin@123"}'
$login = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$token = $login.data.token
Write-Output "Logged in as: $($login.data.user.name) (Role: $($login.data.user.role))"

$headers = @{ Authorization = "Bearer $token" }
$inventory = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/reports/inventory" -Method Get -Headers $headers
Write-Output "Inventory: $($inventory.data | ConvertTo-Json -Compress)"

$assets = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/assets?limit=3" -Method Get -Headers $headers
Write-Output "Assets count: $($assets.meta.total)"

$employees = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users" -Method Get -Headers $headers
Write-Output "Employees with active assets: $($employees.data | ConvertTo-Json -Compress)"
