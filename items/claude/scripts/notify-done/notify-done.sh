#!/bin/bash
# Voice notification when Claude finishes responding
# Uses Windows Speech Synthesis on WSL, or spd-say on native Linux
MSG="Yo! Come back."

if command -v powershell.exe &>/dev/null; then
  powershell.exe -Command "Add-Type -AssemblyName System.Speech; \$s = New-Object System.Speech.Synthesis.SpeechSynthesizer; \$s.Speak('$MSG')" &>/dev/null &
elif command -v spd-say &>/dev/null; then
  spd-say "$MSG" &>/dev/null &
elif command -v espeak &>/dev/null; then
  espeak "$MSG" &>/dev/null &
fi
