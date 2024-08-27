# Definir la ruta del ejecutable antiguo y nuevo
$oldExePath = "C:\ScrapeBlocker\ScrapeBlocker.exe"
$newExePath = "C:\ScrapeBlocker\ScrapeBlocker new.exe"

# Forzar el cierre del proceso ScrapeBlocker.exe
Write-Output "Intentando finalizar ScrapeBlocker.exe..."
Stop-Process -Name "ScrapeBlocker" -Force -ErrorAction SilentlyContinue

# Esperar unos segundos para asegurar que el proceso se haya cerrado
Start-Sleep -Seconds 5

# Verificar si ScrapeBlocker.exe sigue en ejecución
$processes = Get-Process | Where-Object { $_.Name -eq "ScrapeBlocker" }
if ($processes) {
    Write-Output "ScrapeBlocker.exe sigue en ejecución. Saliendo..."
    exit 1
}

# Renombrar el nuevo ejecutable
Write-Output "Renombrando ScrapeBlocker new.exe a ScrapeBlocker.exe..."
Rename-Item -Path $newExePath -NewName "ScrapeBlocker.exe" -ErrorAction Stop

# Reiniciar la aplicación
Write-Output "Reiniciando ScrapeBlocker.exe..."
Start-Process -FilePath "C:\ScrapeBlocker\ScrapeBlocker.exe"

Write-Output "Actualización completada exitosamente."
