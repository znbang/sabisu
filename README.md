# Windows Service Wrapper

## Configuration
wrapper.toml
```toml
[service]
name="SpringApp"
display_name="SpringApp"
description="A Spring Boot application"
start_type="auto" # auto, manual, disabled
interactive=false # true, false
exec_retry=true # true, false
exec_max_retry=5

[exec]
command="c:\\winapp\\graalvm-jdk-21.0.1+12.1\\bin\\java"
args=[
    "-Xms1024m",
    "-Xmx1024m",
    "-Xss256k",
    "-jar", "server.jar",
]
envs=[]

[log]
path="wrapper.log"
max_backup=3
max_size=5 # MB
```

## Run
Running in console:
```sh
sabisu -c wrapper.toml
```
Installing service:
```sh
sabisu -i wrapper.toml
```
Uninstall service:
```sh
sabisu -r wrapper.toml
```