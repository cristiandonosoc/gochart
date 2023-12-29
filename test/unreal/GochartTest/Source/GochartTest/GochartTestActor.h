#pragma once

#include <CoreMinimal.h>
#include <GameFramework/Actor.h>

#include "GochartTestActor.forward_declare.h"

#include "GochartTestActor.generated.h"

UENUM()
enum class EActuatorType : uint8
{
    None,
    Translation,
    Rotation,
    Scale,
};

UCLASS()
class GOCHARTTEST_API AGochartTestActor : public AActor
{
    GENERATED_BODY()

public:
    // Sets default values for this actor's properties
    AGochartTestActor();
    ~AGochartTestActor();

    // Called every frame
    virtual void Tick(float DeltaTime) override;

    UFUNCTION(BlueprintCallable, Category = "Gochart")
    void ToggleActuator(EActuatorType actuator);

    // Delegates - begin
    void StateMainActuator_CaptureUpdate(float dt);
    void StateTranslationActive_Enter();
    void StateTranslationActive_CaptureUpdate(float dt);
    void StateRotationActive_Enter();
    void StateRotationActive_CaptureUpdate(float dt);
    void StateScaleActive_Enter();
    void StateScaleActive_CaptureUpdate(float dt);
    // Delegates - end

protected:
    // Called when the game starts or when spawned
    virtual void BeginPlay() override;
    virtual void EndPlay(const EEndPlayReason::Type EndPlayReason) override;

private:
    using GochartType = gochart::GochartTest::Gochart<AGochartTestActor>;
    friend GochartType;
    TPimplPtr<GochartType> Gochart;


    UPROPERTY(EditAnywhere)
    FVector Speed = FVector(10.0f, 0.0f, 0.0f);

    UPROPERTY(EditAnywhere)
    float Radius = 100.0f;

    FVector TranslationPoint;
    float TranslationAngle = 0.0f;
    float RotationAngle = 0.0f;
    float Scale = 1.0f;
    bool ScalingUp = true;
};