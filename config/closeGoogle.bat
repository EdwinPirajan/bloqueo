@echo off
:: Obtiene el nombre del usuario actual
for /F "tokens=*" %%i in ('whoami') do set CURRENT_USER=%%i

:: Verifica si el usuario actual es .\administrador
if "%CURRENT_USER%"==".\administrador" (
    echo No se cerrarán los procesos ya que el usuario actual es .\administrador
    exit
)

:: Cierra todos los procesos de chrome
taskkill /IM chrome.exe /F

:: Cierra todos los procesos adicionales
taskkill /IM version.exe /F
taskkill /IM javaw.exe /F
taskkill /IM Batch.exe /F
taskkill /IM Control.exe /F
taskkill /IM Clientes.exe /F
taskkill /IM ATMAdmin.exe /F
taskkill /IM ATMCompensacion.exe /F
taskkill /IM ATMMonitor.exe /F
taskkill /IM ATMPersonaliza.exe /F
taskkill /IM PINPAD.exe /F
taskkill /IM Bonos.exe /F
taskkill /IM cartera.exe /F
taskkill /IM COBISCorp.eCOBIS.COBISExplorer.CommunicationManager.exe /F
taskkill /IM COBISExplorerApplicationsRemover /F
taskkill /IM Cce.exe /F
taskkill /IM Cci.exe /F
taskkill /IM Cde.exe /F
taskkill /IM Cdi.exe /F
taskkill /IM Ceadmin.exe /F
taskkill /IM Corresp.exe /F
taskkill /IM Grb-gra.exe /F
taskkill /IM Stb.exe /F
taskkill /IM Tre-trr.exe /F
taskkill /IM Cobconta.exe /F
taskkill /IM cobconci.exe /F
taskkill /IM cobpresu.exe /F
taskkill /IM COBRANZA.exe /F
taskkill /IM Admcred.exe /F
taskkill /IM Tramites.exe /F
taskkill /IM Buzon.exe /F
taskkill /IM Camara.exe /F
taskkill /IM person.exe /F
taskkill /IM prude-po.exe /F
taskkill /IM tadmin.exe /F
taskkill /IM tarifario.exe /F
taskkill /IM sit.exe /F
taskkill /IM af.exe /F
taskkill /IM brp.exe /F
taskkill /IM cxc.exe /F
taskkill /IM cxp.exe /F
taskkill /IM SB.exe /F
taskkill /IM Rechazos.exe /F
taskkill /IM Reportvb5.exe /F
taskkill /IM Depadmin.exe /F
taskkill /IM Depopera.exe /F
taskkill /IM PEB.exe /F
taskkill /IM garantia.exe /F
taskkill /IM Firmas.exe /F
taskkill /IM HerramientaCuadre.exe /F
taskkill /IM vrcAgrario.exe /F

exit
