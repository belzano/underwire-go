go build .
oapi-codegen-client-ue.exe -spec="../jij-service/api/openapi.yaml" -tmpl=templates -out="./generated"

REM go run .\main.go -spec="../jij-service/api/openapi.yaml" -tmpl="templates" -out="./generated"

pause
