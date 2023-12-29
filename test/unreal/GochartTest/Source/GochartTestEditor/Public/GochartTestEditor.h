#pragma once

#include "CoreMinimal.h"

class FGochartTestEditorModule : public IModuleInterface
{
public:
    virtual void StartupModule() override;
    virtual void ShutdownModule() override;
};
