Write-Output "--- Starting End-to-End API Test Suite ---"

# 1. Health Check
$health = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get
Write-Output "1. Health: $($health.message)"

# 2. Login as IT Admin
$adminLogin = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post `
    -Body '{"email":"it.admin@company.com","password":"Admin@123"}' -ContentType "application/json"
$itToken = $adminLogin.data.token
Write-Output "2. IT Admin Login: OK ($($adminLogin.data.user.name))"

# 3. Login as HR Manager
$hrLogin = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post `
    -Body '{"email":"hr.manager@company.com","password":"HR@123"}' -ContentType "application/json"
$hrToken = $hrLogin.data.token
Write-Output "3. HR Manager Login: OK ($($hrLogin.data.user.name))"

# 4. Login as Employee (Jane Smith)
$empLogin = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post `
    -Body '{"email":"jane.smith@company.com","password":"Employee@123"}' -ContentType "application/json"
$empToken = $empLogin.data.token
Write-Output "4. Employee Login: OK ($($empLogin.data.user.name))"

# 5. Inventory Summary (IT Admin)
$inv = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/reports/inventory" -Method Get `
    -Headers @{ Authorization = "Bearer $itToken" }
Write-Output "5. Inventory: Total Assets=$($inv.data.total_assets), Valuation=`$$($inv.data.total_valuation)"

# 6. Employee Check-In
$checkIn = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/attendance/check-in" -Method Post `
    -Headers @{ Authorization = "Bearer $empToken" } `
    -Body '{"status":"present","notes":"Working in office pod 4"}' -ContentType "application/json"
Write-Output "6. Employee Check-In: OK (Date: $($checkIn.data.date), Status: $($checkIn.data.status))"

# 7. Employee Check-Out
$checkOut = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/attendance/check-out" -Method Post `
    -Headers @{ Authorization = "Bearer $empToken" } `
    -Body '{"notes":"All design sprints completed"}' -ContentType "application/json"
Write-Output "7. Employee Check-Out: OK (Work Hours: $($checkOut.data.work_hours))"

# 8. HR Attendance Report
$attReport = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/attendance/report" -Method Get `
    -Headers @{ Authorization = "Bearer $hrToken" }
Write-Output "8. HR Attendance Report: Found $($attReport.meta.total) total logs"

# 9. Employee Submit Leave Request
$leaveReq = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/leaves" -Method Post `
    -Headers @{ Authorization = "Bearer $empToken" } `
    -Body '{"leave_type":"casual","start_date":"2026-12-01T00:00:00Z","end_date":"2026-12-03T00:00:00Z","reason":"Personal family trip"}' `
    -ContentType "application/json"
$leaveId = $leaveReq.data.id
Write-Output "9. Employee Leave Submission: OK (ID: $leaveId, Type: $($leaveReq.data.leave_type))"

# 10. HR Approve Leave Request
$approveLeave = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/leaves/$leaveId/status" -Method Put `
    -Headers @{ Authorization = "Bearer $hrToken" } `
    -Body '{"status":"approved","review_notes":"Approved by HR Manager Sarah Connor"}' `
    -ContentType "application/json"
Write-Output "10. HR Leave Approval: OK (Status: $($approveLeave.data.status))"

Write-Output "--- All API End-to-End Tests Succeeded ---"
