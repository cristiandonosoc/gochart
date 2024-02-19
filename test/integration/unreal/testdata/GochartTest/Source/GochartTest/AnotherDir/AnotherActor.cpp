#include <GochartTest/AnotherDir/AnotherActor.h>

#include "Math/UnrealMathUtility.h"

#include "AnotherDir/Counter.gochart.define.h"

AAnotherActor::AAnotherActor() 
{
    // Set this actor to call Tick() every frame.  You can turn this off to improve performance if
    // you don't need it.
    PrimaryActorTick.bCanEverTick = true;
}

void AAnotherActor::BeginPlay()
{ 
	Super::BeginPlay();

	Count = 0;
    LastUpdate = 0;

	Gochart = MakePimpl<GochartType>(this);
    Gochart->Activate();
}

void AAnotherActor::EndPlay(const EEndPlayReason::Type EndPlayReason) 
{ 
	Gochart->Deactivate(); 
	Gochart.Reset();

	Super::EndPlay(EndPlayReason);
}

void AAnotherActor::Tick(float dt) 
{ 
	constexpr float kTimeToAdd = 3.0f;
	LastUpdate += dt;
	if (LastUpdate < kTimeToAdd)
	{
        return;
	}

    LastUpdate -= kTimeToAdd;
    Gochart->TriggerAdd(FMath::RandRange(1, 5));
}

void AAnotherActor::StateRoot_CaptureAdd(int number) { 
	Count += number; 
	GEngine->AddOnScreenDebugMessage(
        0, 2, FColor::Red, FString::Printf(TEXT("Number: %d, Count: %d"), number, Count));
}

