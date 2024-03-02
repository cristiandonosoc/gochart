#include <NestedDirectory/NestedModule/Some/Deep/Hierarchy/NestedActor.h>

#include <Some/Deep/Hierarchy/NestedActorFlow.gochart.define.h>

#include "Math/UnrealMathUtility.h"

ANestedActor::ANestedActor() : Gochart(this)
{
    // Set this actor to call Tick() every frame.  You can turn this off to improve performance if
    // you don't need it.
    PrimaryActorTick.bCanEverTick = true;
}

void ANestedActor::BeginPlay()
{
    Super::BeginPlay();
    Gochart.Activate();
}

void ANestedActor::EndPlay(const EEndPlayReason::Type EndPlayReason) { Gochart.Deactivate(); }

void ANestedActor::Tick(float dt) { Gochart.TriggerUpdate(dt); }

void ANestedActor::StateRoot_CaptureUpdate(float dt)
{
    constexpr float kTimeToAdd = 1.0f;
    LastUpdate += dt;
    TotalTime += dt;
    if (LastUpdate < kTimeToAdd)
    {
        return;
    }

    FString msg = FString::Printf(TEXT("NESTED ACTOR: %.3f"), TotalTime);
    GEngine->AddOnScreenDebugMessage(0, 2, FColorList::LightBlue, *msg);
    LastUpdate = 0;
}
