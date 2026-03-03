from conan import ConanFile


class TestConan(ConanFile):
    name = "cli-test-subdir-package"
    version = "1.0.0"
    requires = "zlib/1.3.1"

    def build(self):
        self.output.info("Building test package from subdirectory")

    def package(self):
        pass
