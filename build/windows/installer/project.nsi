Unicode true

####
## PassDepot NSIS installer (Wails).
## Shortcut options are checkboxes on the Finish page (not a Components page).
####

!include "wails_tools.nsh"
!include "LogicLib.nsh"

VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

ManifestDPIAware true

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING
!define MUI_LANGDLL_ALLLANGUAGES
!define MUI_FINISHPAGE_RUN
!define MUI_FINISHPAGE_RUN_TEXT "$(CHK_STARTMENU)"
!define MUI_FINISHPAGE_RUN_FUNCTION CreateStartMenuShortcut
!define MUI_FINISHPAGE_SHOWREADME
!define MUI_FINISHPAGE_SHOWREADME_TEXT "$(CHK_DESKTOP)"
!define MUI_FINISHPAGE_SHOWREADME_FUNCTION CreateDesktopShortcut

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"
!insertmacro MUI_LANGUAGE "Russian"

LangString CHK_STARTMENU ${LANG_ENGLISH} "Create Start Menu shortcut"
LangString CHK_STARTMENU ${LANG_RUSSIAN} "Ярлык в меню Пуск"
LangString CHK_DESKTOP ${LANG_ENGLISH} "Create Desktop shortcut"
LangString CHK_DESKTOP ${LANG_RUSSIAN} "Ярлык на рабочий стол"

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"
InstallDir "$PROGRAMFILES64\${INFO_PRODUCTNAME}"
ShowInstDetails show

Function .onInit
   !insertmacro wails.checkArchitecture
   !insertmacro MUI_LANGDLL_DISPLAY
FunctionEnd

Function CreateStartMenuShortcut
    SetShellVarContext all
    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
FunctionEnd

Function CreateDesktopShortcut
    SetShellVarContext current
    CreateShortcut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
FunctionEnd

Section
    !insertmacro wails.setShellContext

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    !insertmacro wails.writeUninstaller
SectionEnd

Section "uninstall"
    SetShellVarContext current

    # Stop a running instance so files/creds can be removed.
    ExecWait 'taskkill /IM PassDepot.exe /F'
    Sleep 500

    # Cleanup AppData + Credential Manager PATs while the binary still exists.
    IfFileExists "$INSTDIR\${PRODUCT_EXECUTABLE}" 0 skip_cleanup
        ExecWait '"$INSTDIR\${PRODUCT_EXECUTABLE}" --uninstall-cleanup'
    skip_cleanup:

    # Belt-and-suspenders if cleanup binary failed / was already gone.
    RMDir /r "$APPDATA\PassDepot"
    RMDir /r "$APPDATA\PassDepot.exe"

    RMDir /r $INSTDIR

    SetShellVarContext all
    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    SetShellVarContext current
    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller
SectionEnd
