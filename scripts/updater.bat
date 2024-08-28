@echo off
timeout /t 5 /nobreak
taskkill /f /im ScrapeBlocker.exe
timeout /t 2 /nobreak
xcopy /y "update\*" ".\"
start "" "ScrapeBlocker.exe"
