package nuget

const ConfigFileTemplate = `<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <add key="Artifactory" value="%s" />
  </packageSources>
  <packageSourceCredentials>
    <Artifactory>
      <add key="Username" value="%s" />
      <add key="ClearTextPassword" value="%s" />
    </Artifactory>
  </packageSourceCredentials>
</configuration>`
