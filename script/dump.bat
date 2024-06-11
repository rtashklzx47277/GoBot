setlocal

set CONTAINER=mysql
set DB_USER=root
set DB_PASSWORD=root
set DB_NAME=mydb
set NOW=%date:~0,4%%date:~5,2%%date:~8,2%
set FILENAME=C:/Users/Derek/Downloads/Workspace/GoBot/backup/%NOW%.sql

docker exec %CONTAINER% sh -c "mysqldump -u %DB_USER% -p%DB_PASSWORD% %DB_NAME%" > %FILENAME%

endlocal
