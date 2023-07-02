@echo off
set pwd=%~dp0
set root=%pwd%\..\..

pushd %pwd%

echo "SCRIPT DIR=%pwd%"
echo "PROJECT ROOT DIR=%root%"

:: Move any old files to a backup location.
if exist statechart.generated.h (
	move statechart.generated.h _statechart.generated.h.BACKUP
)
if exist statechart.generated.cpp (
	move statechart.generated.cpp _statechart.generated.cpp.BACKUP
)

:: First generate the statechart files.
go run %root%\cmd\gochart %root%\pkg\ir\testdata\simple.yaml statechart.generated.h statechart.generated.cpp || goto ERROR

:: Then compile and run the generated cpp case.
bazelisk run ":full_flow" || goto ERROR

goto DONE

:ERROR
echo "ERROR OCURRED"

:DONE
popd
