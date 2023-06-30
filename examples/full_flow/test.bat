@echo off
set pwd=%~dp0
set root=%pwd%\..\..

pushd %pwd%

echo "SCRIPT DIR=%pwd%"
echo "PROJECT ROOT DIR=%root%"

:: Move any old files to a backup location.
move statechart.generated.h _statechart.generated.h.BACKUP
move statechart.generated.cpp _statechart.generated.cpp.BACKUP

:: First generate the statechart files.
go run %root%\cmd\gochart %root%\examples\frontends\yaml\simple.yaml statechart.generated.h statechart.generated.cpp

:: Then compile and run the generated cpp case.
bazelisk run ":full_flow"

popd
