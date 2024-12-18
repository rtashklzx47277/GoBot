@echo off

setlocal

set CONTAINER=mysql
set DB_USER=root
set DB_PASSWORD=root
set DB_NAME=mydb
set NOW=%date:~0,4%%date:~5,2%%date:~8,2%
set FOLDERNAME="C:\Users\Derek\Downloads\Workspace\GoBot\backup"
set FILENAME=%NOW%.sql

docker exec %CONTAINER% sh -c "mysqldump -u %DB_USER% -p%DB_PASSWORD% %DB_NAME%" > %FOLDERNAME%\%FILENAME%

FORFILES /P "%FOLDERNAME%" /M *.sql /D -7 /C "cmd /c del @path"

endlocal
