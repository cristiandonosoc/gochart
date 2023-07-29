@echo off
set pwd=%~dp0
set root=%pwd%\..\..

pushd %pwd%

echo "SCRIPT DIR=%pwd%"
echo "PROJECT ROOT DIR=%root%"

:: Just compile and run the generated cpp case.
bazelisk run ":full_flow" || goto ERROR

goto DONE

:ERROR
echo "ERROR OCURRED"

:DONE
popd
