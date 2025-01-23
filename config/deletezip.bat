@echo off
REM Verifica si el archivo existe
if exist "C:\ScrapeBlocker\scrapeblocker.zip" (
    echo Eliminando scrapeblocker.zip...
    del /f "C:\ScrapeBlocker\scrapeblocker.zip"
    echo Archivo eliminado.
) else (
    echo El archivo scrapeblocker.zip no existe.
)
pause
