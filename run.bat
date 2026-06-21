@echo off
cd /d "%~dp0"
if exist wxtrans.exe (
    start "" wxtrans.exe
) else (
    echo wxtrans.exe not found, running with go run ...
    go run .
)
