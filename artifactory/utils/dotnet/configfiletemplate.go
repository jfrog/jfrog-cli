package dotnet

const ConfigFileTemplate = `<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
  </packageSources>
  <packageSourceCredentials>
  </packageSourceCredentials>
</configuration>`

const ConfigFileFormat = `<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <add key="JFrogCli" value="%s" />
  </packageSources>
  <packageSourceCredentials>
    <JFrogCli>
      <add key="Username" value="%s" />
      <add key="ClearTextPassword" value="%s" />
    </JFrogCli>
  </packageSourceCredentials>
</configuration>`
