@echo off

:: Finalizar todos los procesos ScrapeBlocker.exe
taskkill /F /IM ScrapeBlocker.exe

:: Esperar un momento para asegurarse de que los procesos se cierren
timeout /t 3 /nobreak > nul

:: Eliminar el acceso directo ScrapeBlocker.exe en la carpeta C:\ScrapeBlocker
del "C:\ScrapeBlocker\ScrapeBlocker.exe" /Q

:: Confirmar que el proceso ha terminado y el acceso directo ha sido eliminado
echo Todos los procesos ScrapeBlocker.exe han sido finalizados y el acceso directo ha sido eliminado.
pause
