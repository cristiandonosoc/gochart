#pragma once

#include "CoreMinimal.h"

#include "AnotherDir/Counter.gochart.forward_declare.h"

#include "AnotherActor.generated.h"

UCLASS()
class GOCHARTTEST_API AAnotherActor : public AActor
{
    GENERATED_BODY()

protected:
    AAnotherActor();

    // Called when the game starts or when spawned
    virtual void BeginPlay() override;
    virtual void EndPlay(const EEndPlayReason::Type EndPlayReason) override;

    void Tick(float dt) override;

 private:
    using GochartType = gochart::Counter::Gochart<AAnotherActor>;
    friend GochartType;
    TPimplPtr<GochartType> Gochart;

    int Count = 0;
    float LastUpdate = 0;

    void StateRoot_CaptureAdd(int number);
};
