@echo off
setlocal

REM =========================
REM 1. 路径配置（统一改这里）
REM =========================
set ROOT=E:\go\tool\.sys
set VIPS=%ROOT%\vips

REM =========================
REM 2. 生成 vips.pc（你已有）
REM =========================
go run gen_pc.go

REM =========================
REM 3. 关键环境变量（核心）
REM =========================

REM 指定 pkg-config
set PKG_CONFIG=%ROOT%\pkg-config.exe

REM ❗最关键：必须指向 vips 的 pkgconfig（里面包含 glib-2.0.pc）
set PKG_CONFIG_PATH=%VIPS%\lib\pkgconfig

REM 可选：确保 DLL 能找到（运行时用）
set PATH=%VIPS%\bin;%PATH%

REM =========================
REM 4. 调试输出（建议保留）
REM =========================
echo ==== pkg-config check ====
%PKG_CONFIG% --cflags vips
%PKG_CONFIG% --libs vips

echo ==========================

REM =========================
REM 5. 开始编译
REM =========================
go build -ldflags "-s -w" -o toolsa.exe

echo.
echo ==== BUILD DONE ====
pause
