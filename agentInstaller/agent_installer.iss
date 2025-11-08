; agent_installer.iss
; Установщик для agent.exe с параметрами и шедулером

[Setup]
AppName=Agent
AppVersion=1.0
DefaultDirName={pf}\Agent
DisableProgramGroupPage=yes
UninstallDisplayIcon={app}\agent.exe
OutputBaseFilename=agent_installer
Compression=lzma
SolidCompression=yes
PrivilegesRequired=admin

[Files]
Source: "agent.exe"; DestDir: "{app}"; Flags: ignoreversion

[Code]
var
  LogPath, ProxyId: String;

procedure InitializeWizard;
begin
  MsgBox('Добро пожаловать в установку Agent!', mbInformation, MB_OK);

  LogPath := InputBox('Путь к логам', 
    'Введите полный путь к файлу логов (например, C:\Agent\logs\agent.log):', '');
  ProxyId := InputBox('Proxy ID', 'Введите идентификатор Proxy ID:', '');
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  Params, Cmd: String;
begin
  if CurStep = ssPostInstall then
  begin
    Params := '--log-path=' + LogPath + 
              ' --api-host=http://135.181.144.163:8080' + 
              ' --proxy-id=' + ProxyId;

    Cmd := '/create /tn "AgentDailyRun" /tr "' + 
           '"' + ExpandConstant('{app}\agent.exe') + '" ' + Params + 
           '" /sc daily /st 00:00 /ru SYSTEM /f';

    Exec('schtasks', Cmd, '', SW_HIDE, ewWaitUntilTerminated, nil);

    MsgBox('Установка завершена! Задача AgentDailyRun создана в планировщике.', mbInformation, MB_OK);
  end;
end;
