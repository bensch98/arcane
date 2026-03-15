#!/bin/bash
# Windows toast notification when Claude finishes responding
TITLE="Claude Code"
MSG="Task is done. Claude is waiting for you."

if command -v powershell.exe &>/dev/null; then
  powershell.exe -Command "
    [Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > \$null
    \$template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
    \$textNodes = \$template.GetElementsByTagName('text')
    \$textNodes.Item(0).AppendChild(\$template.CreateTextNode('$TITLE')) > \$null
    \$textNodes.Item(1).AppendChild(\$template.CreateTextNode('$MSG')) > \$null
    \$toast = [Windows.UI.Notifications.ToastNotification]::new(\$template)
    [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('$TITLE').Show(\$toast)
  " &>/dev/null &
fi
