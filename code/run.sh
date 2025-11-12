
#!/bin/bash

# Бесконечный цикл для перезапуска приложения
while true; do
    # Сборка и запуск
    go build -o proxychecker . && ./proxychecker
    
    # Выводим сообщение о перезапуске
    echo "Application exited. Restarting in 1 second..."
    sleep 1
done
F