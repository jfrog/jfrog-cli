package example;

/**
 * Simple example class for the gradle test project.
 * Note: We intentionally don't use JUnit here because dependency resolution
 * may not work when the JFrog Gradle plugin configures the resolver to use
 * the test Artifactory repository (which may not have JUnit cached).
 */
public class Example {
    public boolean isTrue() {
        return true;
    }
    
    public static void main(String[] args) {
        Example example = new Example();
        System.out.println("Example.isTrue() = " + example.isTrue());
    }
}

