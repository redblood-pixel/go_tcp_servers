Для сборки образа локально
```
docker build -t client-app .
```

Для запуска серверов из корневой папки проекта(со службами)
```
bash run.sh
```

Для подключения через Docker-клиент к первому серверу
```
docker run --add-host=host.docker.internal:host-gateway \
           -e SERVER1_ADDR=host.docker.internal:5001 \
           -e SERVER2_ADDR=host.docker.internal:5002 \
           -e SERVER_CHOICE=1 \
           -e TEXT_COLOR=red \
           client-app
```

Ко второму серверу
```
docker run --add-host=host.docker.internal:host-gateway \
           -e SERVER1_ADDR=host.docker.internal:5001 \
           -e SERVER2_ADDR=host.docker.internal:5002 \
           -e SERVER_CHOICE=2 \
           -e PARAM_TYPE=1 \
           -e PARAMETER=2 \
           client-app
```
