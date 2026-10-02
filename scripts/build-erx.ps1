$ErrorActionPreference = 'Stop'

$savedEnvironment = @{}
foreach ($name in @('CGO_ENABLED', 'GOOS', 'GOARCH', 'GOMIPS')) {
	$savedEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
}

try {
	$env:CGO_ENABLED = '0'
	$env:GOOS = 'linux'
	$env:GOARCH = 'mipsle'
	$env:GOMIPS = 'softfloat'

	New-Item -ItemType Directory -Force -Path dist | Out-Null
	go build -trimpath -ldflags '-s -w' -o dist/edge-sensor-linux-mipsle ./cmd/edge-sensor
	if ($LASTEXITCODE -ne 0) {
		throw "ER-X build failed with exit code $LASTEXITCODE"
	}
	Write-Host 'Built dist/edge-sensor-linux-mipsle'
} finally {
	foreach ($name in $savedEnvironment.Keys) {
		[Environment]::SetEnvironmentVariable($name, $savedEnvironment[$name], 'Process')
	}
}