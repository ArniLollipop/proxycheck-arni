
1. Скопируй app.exe Например в корень C:
2. Создаем таску 
3. Запускаем командную строку  от имени администратора и выполняем 
schtasks /Create /TN "AgentDailyRun" /TR "\"C:\agent.exe\" --log-path=\"C:\Agent\logs" --api-host=http://135.181.144.163:8080" /SC DAILY /ST 00:00 /RL HIGHEST /F /RU SYSTEM

Где --log-path - путь к папке логов 

Готово Теперь раз в сутки будет проверять логи 
