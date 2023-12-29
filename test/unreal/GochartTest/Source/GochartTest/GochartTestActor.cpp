#include <GochartTest/GochartTestActor.h>

#include "GochartTestActor.define.h"

namespace
{

FVector
RotateAroundAxis(const FVector& origin, const FVector& point, const FVector& axis, float angle)
{
    FVector translatedPoint = point - origin;
    FQuat rotation = FQuat(axis, FMath::DegreesToRadians(angle));
    FVector rotatedPoint = rotation.RotateVector(translatedPoint);

    return origin + rotatedPoint;
}

} // namespace

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

    Gochart = MakePimpl<GochartType>(this);
    Gochart->Activate();
}

void AGochartTestActor::EndPlay(const EEndPlayReason::Type EndPlayReason)
{
    Gochart->Deactivate();
    Gochart.Reset();

    Super::EndPlay(EndPlayReason);
}

// Called every frame
void AGochartTestActor::Tick(float DeltaTime)
{
    Super::Tick(DeltaTime);

    Gochart->TriggerUpdate(DeltaTime);
}

void AGochartTestActor::StateMainActuator_CaptureUpdate(float dt)
{
    DrawDebugSphere(GetWorld(), GetActorLocation(), 50.0f, 16, FColor::Blue);
}

void AGochartTestActor::StateTranslationActive_Enter()
{
    TranslationAngle = 0.0f;
    TranslationPoint = GetActorLocation() + Radius * GetActorRightVector();
}

void AGochartTestActor::StateTranslationActive_CaptureUpdate(float dt)
{
    constexpr float kDegreesPerSecond = 90.0f;
    TranslationAngle += kDegreesPerSecond * dt;
    if (TranslationAngle > 360.0f)
    {
        TranslationAngle -= 360.0f;
    }

    FVector location = RotateAroundAxis(
        GetActorLocation(), TranslationPoint, GetActorUpVector(), TranslationAngle);

    if (auto* mesh = FindComponentByClass<UMeshComponent>())
    {
        mesh->SetWorldLocation(location);
    }

    GEngine->AddOnScreenDebugMessage(
        __LINE__,
        INDEX_NONE,
        FColor::Blue,
        FString::Printf(TEXT("TranslationAngle: %f"), TranslationAngle));
}

void AGochartTestActor::StateRotationActive_Enter() { SetActorRotation(FRotator::ZeroRotator); }

void AGochartTestActor::StateRotationActive_CaptureUpdate(float dt)
{
    constexpr float kDegreesPerSecond = 90.0f;
    RotationAngle += kDegreesPerSecond * dt;
    if (RotationAngle > 360.0f)
    {
        RotationAngle -= 360.0f;
    }

    float rotation = kDegreesPerSecond * dt;
    FRotator increment(0.0f, rotation, 0.0f);
    AddActorLocalRotation(increment);

    GEngine->AddOnScreenDebugMessage(__LINE__,
                                     INDEX_NONE,
                                     FColor::Blue,
                                     FString::Printf(TEXT("RotationAngle: %f"), TranslationAngle));
}

void AGochartTestActor::StateScaleActive_Enter()
{
    Scale = 1.0f;
    ScalingUp = true;
}

void AGochartTestActor::StateScaleActive_CaptureUpdate(float dt)
{
    constexpr float kScaleChangePerSecond = 0.25f;
    float sign = ScalingUp ? 1.0f : -1.0f;
    Scale += sign * kScaleChangePerSecond * dt;

    if (Scale > 1.75f || Scale < 0.25f)
    {
        ScalingUp = !ScalingUp;
    }

    SetActorScale3D(FVector(Scale));

    GEngine->AddOnScreenDebugMessage(
        __LINE__, INDEX_NONE, FColor::Blue, FString::Printf(TEXT("Scale: %f"), Scale));
}

void AGochartTestActor::ToggleActuator(EActuatorType actuator)
{
    if (!Gochart.IsValid())
    {
        return;
    }

    Gochart->TriggerToggleActuator(actuator);
}