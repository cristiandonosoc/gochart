@echo off
set pwd=%~dp0
set root=%pwd%\..\..

echo "SCRIPT DIR=%pwd%"
echo "PROJECT ROOT DIR=%root%"

go run %root%\cmd\gochart %root%\examples\frontends\yaml\simple.yaml
