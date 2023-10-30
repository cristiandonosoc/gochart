#include <GochartTest/GochartTestActor.h>

// Sets default values
AGochartTestActor::AGochartTestActor()
{
    // Set this actor to call Tick() every frame.  You can turn this off to improve performance if
    // you don't need it.
    PrimaryActorTick.bCanEverTick = true;
}

// Called when the game starts or when spawned
void AGochartTestActor::BeginPlay() { Super::BeginPlay(); }

// Called every frame
void AGochartTestActor::Tick(float DeltaTime)
{
    Super::Tick(DeltaTime);

    FVector location = GetActorLocation();
    location += Speed;
    SetActorLocation(location);
}