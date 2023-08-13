@echo off
set pwd=%~dp0
set root=%pwd%..\..

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

@echo on
:: First generate the statechart files.
go run %root%\cmd\gochart --verbose --frontend=gochart --header-output=%pwd%\statechart.generated.h --body-output=%pwd%\statechart.generated.cpp %pwd%\full_flow.gochart || goto ERROR

:: Then compile and run the generated cpp case.
bazelisk run ":full_flow" || goto ERROR


goto DONE

:ERROR
@echo off
echo "ERROR OCURRED"

:DONE
@echo off
popd
