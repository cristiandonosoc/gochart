#pragma once

#include "IDetailCustomization.h"

class GOCHARTTESTEDITOR_API FGochartTestActorDetailsCustomization : public IDetailCustomization
{
public:
    static TSharedRef<IDetailCustomization> MakeInstance();

    virtual void CustomizeDetails(IDetailLayoutBuilder& detail_builder) override;
};
