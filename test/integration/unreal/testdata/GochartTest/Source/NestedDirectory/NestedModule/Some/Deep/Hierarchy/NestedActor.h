#pragma once

#include "CoreMinimal.h"

#include "NestedActor.generated.h"

UCLASS()
class NESTEDMODULE_API ANestedActor : public AActor
{
    GENERATED_BODY()

protected:
    ANestedActor();

    // Called when the game starts or when spawned
    virtual void BeginPlay() override;
    virtual void EndPlay(const EEndPlayReason::Type EndPlayReason) override;

    void Tick(float dt) override;

private:
    float LastUpdate = 0;
    float TotalTime = 0;
};
