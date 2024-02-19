#include <GochartTestActorDetailsCustomization.h>

#include <DetailCategoryBuilder.h>
#include <DetailLayoutBuilder.h>
#include <DetailWidgetRow.h>
#include <GochartTest/GochartTestActor.h>

namespace
{

void AddGochartButtonRow(IDetailCategoryBuilder& category,
                         TWeakObjectPtr<AGochartTestActor> gochart,
                         EActuatorType actuator,
                         const TCHAR* label)
{
    // clang-format off
    category.AddCustomRow(FText::FromString("Gochart Actuators"))
    .ValueContent()
    [
        SNew(SButton)
        .Text(FText::FromString(label))
        .OnClicked_Lambda([gochart, actuator]
        {
            if (gochart.IsValid())
            {
                gochart->ToggleActuator(actuator);
            }
            return FReply::Handled();
        })
    ];
    // clang-format on
}
} // namespace

TSharedRef<IDetailCustomization> FGochartTestActorDetailsCustomization::MakeInstance()
{
    return MakeShareable(new FGochartTestActorDetailsCustomization);
}

void FGochartTestActorDetailsCustomization::CustomizeDetails(IDetailLayoutBuilder& detail_builder)
{
    TArray<TWeakObjectPtr<UObject>> objects;
    detail_builder.GetObjectsBeingCustomized(objects);

    // Try to get at least one GochartTestActor.
    TWeakObjectPtr<AGochartTestActor> gochart = nullptr;
    for (auto& object : objects)
    {
        if (auto* casted = Cast<AGochartTestActor>(object.Get()))
        {
            gochart = casted;
            break;
        }
    }

    // If we didn't find anything, move on.
    if (!gochart.IsValid())
    {
        return;
    }

    IDetailCategoryBuilder& category = detail_builder.EditCategory("Gochart");
    AddGochartButtonRow(category, gochart, EActuatorType::Translation, TEXT("Toggle Translation"));
    AddGochartButtonRow(category, gochart, EActuatorType::Rotation, TEXT("Toggle Rotation"));
    AddGochartButtonRow(category, gochart, EActuatorType::Scale, TEXT("Toggle Scale"));
}