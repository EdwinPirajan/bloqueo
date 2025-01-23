@echo off
REM Cambiar al directorio del programa
cd C:\ScrapeBlocker

REM Terminar cualquier proceso existente de ScrapeBlocker.exe sin validación
taskkill /F /IM ScrapeBlocker.exe /T

REM Iniciar ScrapeBlocker.exe
start ScrapeBlocker.exe
