#include <GochartTest/GochartTestActor.h>

#include "GochartTestActor.define.h"

// Sets default values
AGochartTestActor::AGochartTestActor()
{
    // Set this actor to call Tick() every frame.  You can turn this off to improve performance if
    // you don't need it.
    PrimaryActorTick.bCanEverTick = true;
}

AGochartTestActor::~AGochartTestActor() = default;

// Called when the game starts or when spawned
void AGochartTestActor::BeginPlay()
{
    Super::BeginPlay();
    Gochart = gochart::GochartTest::Create(this);
    Gochart->Activate();
}

void AGochartTestActor::EndPlay(const EEndPlayReason::Type EndPlayReason)
{
    Gochart->Deactivate();
    Gochart.reset();

    Super::EndPlay(EndPlayReason);
}

// Called every frame
void AGochartTestActor::Tick(float DeltaTime)
{
    Super::Tick(DeltaTime);

    FVector location = GetActorLocation();
    location += Speed;
    SetActorLocation(location);
}