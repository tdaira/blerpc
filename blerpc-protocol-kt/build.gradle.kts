plugins {
    `java-library`
    kotlin("jvm") version "1.9.22"
    id("com.google.protobuf") version "0.9.4"
}

group = "com.blerpc"
version = "0.1.0"

java {
    sourceCompatibility = JavaVersion.VERSION_17
    targetCompatibility = JavaVersion.VERSION_17
}

tasks.withType<org.jetbrains.kotlin.gradle.tasks.KotlinCompile> {
    kotlinOptions.jvmTarget = "17"
}

repositories {
    mavenCentral()
}

protobuf {
    protoc {
        artifact = "com.google.protobuf:protoc:3.25.2"
    }
    generateProtoTasks {
        all().forEach { task ->
            task.builtins {
                named("java") {
                    option("lite")
                }
            }
        }
    }
}

dependencies {
    api("com.google.protobuf:protobuf-javalite:3.25.2")
    testImplementation("junit:junit:4.13.2")
}
