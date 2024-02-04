tell application "Mail"
	set localMailboxes to every mailbox
	set messageCountDisplay to ""

	set everyAccount to every account
	repeat with eachAccount in everyAccount
		set accountMailboxes to every mailbox of eachAccount
		if (count of accountMailboxes) is greater than 0 then
			set messageCountDisplay to messageCountDisplay & name of eachAccount & ": " & my getMessageCountsForMailboxes(accountMailboxes) & "\n"
		end if
	end repeat

	return messageCountDisplay
end tell

on getMessageCountsForMailboxes(theMailboxes)
	-- (list of mailboxes)
	-- returns string
	
	tell application "Mail"
		repeat with eachMailbox in theMailboxes
			set mailboxName to name of eachMailbox
			if mailboxName is equal to "INBOX" then
				return (count of (messages of eachMailbox)) as string
			end if
		end repeat
	end tell
	
	return ""
end getMessageCountsForMailboxes
