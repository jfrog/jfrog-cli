package utils


const GradleInitScript =
`import org.jfrog.gradle.plugin.artifactory.ArtifactoryPlugin
import org.jfrog.gradle.plugin.artifactory.task.ArtifactoryTask

initscript {
    dependencies {
        classpath fileTree('${pluginLibDir}')
    }
}

addListener(new BuildInfoPluginListener())
class BuildInfoPluginListener extends BuildAdapter {

    def void projectsLoaded(Gradle gradle) {
        gradle.startParameter.getProjectProperties().put("build.start", Long.toString(System.currentTimeMillis()))
        Project root = gradle.getRootProject()
        root.logger.debug("Artifactory plugin: projectsEvaluated: ${root.name}")
        if (!"buildSrc".equals(root.name)) {
            root.allprojects {
                apply {
                    apply plugin: ArtifactoryPlugin
                }
            }
        }

        // Set the "archives" configuration to all Artifactory tasks.
        for (Project p : root.getAllprojects()) {
            Task t = p.getTasks().findByName(ArtifactoryTask.ARTIFACTORY_PUBLISH_TASK_NAME)
            if (t != null) {
                ArtifactoryTask task = (ArtifactoryTask)t
                task.setAddArchivesConfigToTask(true)
            }
        }
    }
}
`