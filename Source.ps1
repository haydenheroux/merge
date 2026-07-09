param(
    [string]$Path = ".env"
)

if (-not (Test-Path $Path)) {
    Write-Error "File not found: $Path"
    return
}

Get-Content $Path | ForEach-Object {
    $line = $_.Trim()

    # Skip blank lines and comments
    if (-not $line -or $line.StartsWith("#")) {
        return
    }

    # Match KEY=VALUE
    if ($line -match '^\s*([^=]+?)\s*=\s*(.*)\s*$') {
        $key = $matches[1].Trim()
        $value = $matches[2]

        # Remove surrounding quotes if present
        if (
            ($value.StartsWith('"') -and $value.EndsWith('"')) -or
            ($value.StartsWith("'") -and $value.EndsWith("'"))
        ) {
            $value = $value.Substring(1, $value.Length - 2)
        }

        Set-Item -Path "Env:$key" -Value $value
    }
}