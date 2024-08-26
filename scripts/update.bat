@echo off
timeout /t 5 /nobreak
taskkill /f /im main.exe
timeout /t 2 /nobreak
xcopy /y "update\*" ".\"
start "" "main.exe"
