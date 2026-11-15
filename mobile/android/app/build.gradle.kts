import java.io.File
import java.util.Properties

plugins {
    id("com.android.application")
    // The Flutter Gradle Plugin must be applied after the Android and Kotlin Gradle plugins.
    id("dev.flutter.flutter-gradle-plugin")
}

// Release signing credentials are stored outside version control in
// mobile/android/key.properties (see mobile/android/.gitignore).
val signingPropertiesFile = rootProject.file("key.properties")
val signingProperties = Properties()
val requiredSigningProperties = listOf(
    "storeFile",
    "storePassword",
    "keyAlias",
    "keyPassword",
)
val isReleaseBuildRequested = gradle.startParameter.taskNames.any {
    it.contains("release", ignoreCase = true)
}

if (signingPropertiesFile.exists()) {
    signingPropertiesFile.inputStream().use { signingProperties.load(it) }
}

fun missingSigningProperty(name: String): Nothing = throw GradleException(
    "Release signing is not configured: '$name' is missing or empty in " +
        "${signingPropertiesFile.absolutePath}. Add the required signing values " +
        "to that file (never commit it) before building a release APK."
)

val hasSigningProperties = signingPropertiesFile.exists() &&
    requiredSigningProperties.all { !signingProperties.getProperty(it).isNullOrBlank() }

if (isReleaseBuildRequested) {
    if (!signingPropertiesFile.exists()) {
        throw GradleException(
            "Release signing is not configured: ${signingPropertiesFile.absolutePath} " +
                "was not found. Create it with storeFile, storePassword, keyAlias, " +
                "and keyPassword (never commit it) before building a release APK."
        )
    }
    if (!hasSigningProperties) {
        requiredSigningProperties.firstOrNull {
            signingProperties.getProperty(it).isNullOrBlank()
        }?.let { missingSigningProperty(it) }
    }
}

val releaseStoreFile: File? = if (hasSigningProperties) {
    val rawStoreFile = signingProperties.getProperty("storeFile").trim()
    val candidate = File(rawStoreFile).let { file ->
        if (file.isAbsolute) file else rootProject.file(rawStoreFile)
    }
    if (!candidate.exists()) {
        if (isReleaseBuildRequested) {
            throw GradleException(
                "Release signing is not configured: the keystore referenced by " +
                    "'storeFile' in ${signingPropertiesFile.absolutePath} does not " +
                    "exist at the configured path. Update the path to the existing " +
                    "permanent release keystore; do not generate a replacement keystore."
            )
        }
        null
    } else {
        candidate
    }
} else {
    null
}

android {
    namespace = "com.example.vendorsaathi_mobile"
    compileSdk = flutter.compileSdkVersion
    ndkVersion = flutter.ndkVersion

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    defaultConfig {
        // TODO: Specify your own unique Application ID (https://developer.android.com/studio/build/application-id.html).
        applicationId = "com.example.vendorsaathi_mobile"
        // You can update the following values to match your application needs.
        // For more information, see: https://flutter.dev/to/review-gradle-config.
        minSdk = flutter.minSdkVersion
        targetSdk = flutter.targetSdkVersion
        // Uses the version code from pubspec.yaml. When using split APKs, 1000 * ABI_VERSION
        // is added automatically by Flutter. (https://developer.android.com/studio/build/configure-apk-splits#configure-APK-versions)
        // You can force using the value of versionCode by specifying the `-P force-version-code-ignoring-abi=true`
        // flag during build.
        versionCode = flutter.versionCode
        versionName = flutter.versionName
    }

    signingConfigs {
        if (hasSigningProperties && releaseStoreFile != null) {
            create("release") {
                storeFile = releaseStoreFile
                storePassword = signingProperties.getProperty("storePassword")
                keyAlias = signingProperties.getProperty("keyAlias")
                keyPassword = signingProperties.getProperty("keyPassword")
            }
        }
    }

    buildTypes {
        release {
            if (hasSigningProperties && releaseStoreFile != null) {
                // Always sign release builds with the permanent production keystore.
                signingConfig = signingConfigs.getByName("release")
            }
            // When signing is absent, release-task validation above has already
            // failed with a clear message. Debug builds remain usable.
        }
    }
}

kotlin {
    compilerOptions {
        jvmTarget = org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17
    }
}

flutter {
    source = "../.."
}
