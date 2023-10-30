#pragma once

#include <CoreMinimal.h>
#include <GameFramework/Actor.h>

#include "GochartTestActor.generated.h"

UCLASS()
class GOCHARTTEST_API AGochartTestActor : public AActor
{
    GENERATED_BODY()

public:
    // Sets default values for this actor's properties
    AGochartTestActor();

    UPROPERTY(EditAnywhere)
    FVector Speed = FVector(10.0f, 0.0f, 0.0f);

protected:
    // Called when the game starts or when spawned
    virtual void BeginPlay() override;

public:
    // Called every frame
    virtual void Tick(float DeltaTime) override;
};