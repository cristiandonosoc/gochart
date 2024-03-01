#include <NestedDirectory/NestedModule/Some/Deep/Hierarchy/NestedActor.h>

#include "Math/UnrealMathUtility.h"

ANestedActor::ANestedActor()
{
    // Set this actor to call Tick() every frame.  You can turn this off to improve performance if
    // you don't need it.
    PrimaryActorTick.bCanEverTick = true;
}

void ANestedActor::BeginPlay()
{
	Super::BeginPlay();
}

void ANestedActor::EndPlay(const EEndPlayReason::Type EndPlayReason)
{
}

void ANestedActor::Tick(float dt)
{
	constexpr float kTimeToAdd = 3.0f;
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
