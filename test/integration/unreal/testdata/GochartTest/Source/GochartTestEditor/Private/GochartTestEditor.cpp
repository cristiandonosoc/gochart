#include "GochartTestEditor.h"

#include <GochartTestActorDetailsCustomization.h>
#include <PropertyEditorModule.h>

#define LOCTEXT_NAMESPACE "FGochartTestEditorModule"

void FGochartTestEditorModule::StartupModule()
{
    auto& property_module =
        FModuleManager::LoadModuleChecked<FPropertyEditorModule>("PropertyEditor");
    property_module.RegisterCustomClassLayout(
        "GochartTestActor",
        FOnGetDetailCustomizationInstance::CreateStatic(
            &FGochartTestActorDetailsCustomization::MakeInstance));
}

void FGochartTestEditorModule::ShutdownModule() {}

#undef LOCTEXT_NAMESPACE

IMPLEMENT_MODULE(FGochartTestEditorModule, GochartTestEditor)