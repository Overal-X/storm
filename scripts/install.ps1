$ErrorActionPreference = "Stop"

# Resolve the latest version from GitHub Releases API when no explicit version is provided
$LatestReleaseApi = "https://api.github.com/repos/Overal-X/formatio.storm/releases/latest"

# Get the version from the command line argument or resolve latest
$Version = $args[0]
if (-not $Version) {
    try {
        $ReleaseInfo = Invoke-RestMethod -Uri $LatestReleaseApi -Headers @{ Accept = "application/vnd.github+json" }
        $Version = $ReleaseInfo.tag_name
        if (-not $Version) {
            throw "Missing tag_name in GitHub API response"
        }
    }
    catch {
        Write-Host "Failed to resolve latest release version from GitHub API: $($_.Exception.Message)"
        exit 1
    }
}

# Define the base URL for the release artifacts
$BaseUrl = "https://github.com/Overal-X/formatio.storm/releases/download/$Version"

# Define the file names (adjust these as needed)
$Windows_AMD64 = "storm_Windows_x86_64.zip"
$Windows_ARM64 = "storm_Windows_arm64.zip"

# Determine the architecture
$Arch = [System.Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITECTURE")

# Set the file to download based on architecture
switch ($Arch) {
    "AMD64" {
        $File = $Windows_AMD64
    }
    "ARM64" {
        $File = $Windows_ARM64
    }
    default {
        Write-Host "Unsupported architecture: $Arch"
        exit 1
    }
}

# Define the destination directory
$DestDir = "$env:USERPROFILE\.storm\bin"

# Create the destination directory if it does not exist
if (-not (Test-Path $DestDir)) {
    New-Item -Path $DestDir -ItemType Directory | Out-Null
}

# Download the file
Write-Host "Using version: $Version"
Write-Host "Downloading $File..."
try {
    Invoke-WebRequest -Uri "$BaseUrl/$File" -OutFile "$File"
} catch {
    Write-Error "Failed to download $BaseUrl/$File`: $($_.Exception.Message)"
    exit 1
}

# Extract the downloaded file to the .storm\bin directory
Write-Host "Extracting $File to $DestDir..."
try {
    Expand-Archive -Path $File -DestinationPath $DestDir -Force
} catch {
    Write-Error "Failed to extract ${File}: $($_.Exception.Message)"
    Remove-Item -Path $File -Force -ErrorAction SilentlyContinue
    exit 1
}

# Remove the downloaded file
Remove-Item -Path $File -Force

# Add .storm\bin to the PATH if it's not already there
$PathEntry = "$DestDir"
$CurrentPath = [System.Environment]::GetEnvironmentVariable("PATH", [System.EnvironmentVariableTarget]::User)

if ($CurrentPath -notlike "*$PathEntry*") {
    [System.Environment]::SetEnvironmentVariable("PATH", "$CurrentPath;$PathEntry", [System.EnvironmentVariableTarget]::User)
    Write-Host "Updated PATH to include $DestDir"
} else {
    Write-Host "$DestDir is already in PATH"
}
