cd /d %~dp0
go build -o %cd%/../bin/cnf.exe  %cd%/../cmd/main.go
%cd%/../bin/cnf.exe --configure ../config/config.default.json