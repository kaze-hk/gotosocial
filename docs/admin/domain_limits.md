# Domain Limits

Via the settings panel, GoToSocial allows you to create "domain limits" to modify how your instance handles posts, follow requests, media, and accounts from remote domains.

This gives you the power to "fine-tune" federation with a problematic domain, without having to resort to putting a full [domain block](./domain_blocks.md) in place to cut off federation entirely.

For example, you can use a domain limit to mute all accounts on a given domain except for ones people on your instance follow, and/or to mark all media from a given domain as "sensitive", etc.

!!! tip
    When you create a domain limit, it extends to all subdomains as well, so limiting 'example.com' also limits 'social.example.com'.

You can view, create, and remove domain limits using the [instance admin panel](./settings.md#domain-limits).

Each domain limit has five components that you can tweak to tune federation with a limited domain:

- Content warning
- Media policy
- Follows policy
- Statuses policy
- Accounts policy

## Content Warning

Any text that you set as a content warning will be added to each post from the limited domain as part of the content warning / subject field when that post is viewed using a client app (in the home or public timeline, for example).

If the post already has a content warning, any text set as the limit content warning will be prepended to the existing content warning with a semicolon, so that the existing content warning is not lost.

!!! tip
    Filling the content warning field will also have the effect of marking all posts (and attachments) from the limited domain as sensitive.

Without domain limit content warning (in Pinafore):

![A post from user "big gerald" with some text and an image](../public/domain-limits-unlimited.png)

With domain limit content warning (in Pinafore):

![The same post as above, but it now has a content warning and the media is marked sensitive](../public/domain-limits-content-warning.png)

## Media Policy

You can apply a media policy in order to change whether and how your instance processes media attachments from the limited domain.

For example, you can force all media from the limited domain to be marked sensitive when viewed using a client app. Or you can reject media from the limited domain entirely, which will prevent your instance from downloading, processing, or storing it (this includes attachments, avatars, headers, and emoji).

With a "reject" policy in place, media from the limited domain will still be linked to when a post containing media is viewed via a client app, so that users can still view the media on the remote instance if they choose to:

![A post with some text and a link below the text to the rejected media file](../public/domain-limits-media-rejected.png)

!!! warning
    Setting media policy to "reject" will prevent new media from that domain from being downloaded, but it will not immediately clear all media from the target domain from your instance's storage. Rather, it will be gradually uncached according your instance's [media caching](./media_caching.md) settings, and when it is uncached it will not be automatically recached by your instance.
    
    If you want to *immediately* remove all cached media from a limited domain, and not have your instance download it again, then after you put a "reject" policy in place, you should run a media purge action on the limited domain using the [admin actions part of the settings panel](./settings.md#media).

## Follows Policy

You can apply a follows policy to determine how follows from the limited domain are processed.

Any restrictions put in place only apply to new follows sent to your instance from the moment you apply the policy; existing relationships will not be affected. Also, accounts on your instance will still be able to follow accounts from the limited domain as normal.

If you set this policy to "manual approval", then all follow requests originating from the limited domain will require manual approval, even if they target an "unlocked" account on your instance, ie., an account that would normally not place any restrictions on follows.

If you set this policy to "reject non-mutual", then each follow request originating from the limited domain will be automatically rejected *unless* it is a "follow-back" or "mutual" follow. For example, if user A on this instance already follows or follow-requests user B from the limited domain, user B will be able to send a follow (request) to user A as normal. However, if user A on this instance does *not* already follow or follow-request user B from the limited domain, any attempt by user B to follow user A will be automatically rejected.

If you set this policy to "reject" then any follow requests originating from the limited domain will be instantly and automatically rejected, without creating any notifications.

## Statuses Policy

You can apply a statuses policy to determine if and how statuses (aka posts) from the limited domain are filtered when viewed using a client app. Any filters applied via this policy apply in the contexts `home`, `public`, and `thread`.

If you set this to "warn" then a warn-level filter will be applied to posts from accounts from the limited domain. Users will have to click "show anyway" (depending on their client) in order to show posts filtered in this way. This allows you to add a small level of friction to showing posts from the limited domain.

If you set this to "hide" then a hide-level filter will be applied to posts from accounts from the limited domain. This means that posts from filtered accounts will be hidden from the public/federated timeline, and boosts of and replies by those accounts will also not be shown. However, posts will still be visible when navigating to an account's profile page. This allows you to add a larger level of friction to showing posts from the limited domain.

No action:

![The federated timeline showing a post at the top from user "big gerald" from "fossbros-anonymous.io"](../public/domain-limits-federated.png)

Warn:

![The same view of the federated timeline, but the post from "big gerald" is collapsed behind a "show anyway" dialogue](../public/domain-limits-federated-warn.png)

Hide:

![The same view of the federated timeline, but the post from "big gerald" is now not shown whatsoever](../public/domain-limits-federated-hide.png)

!!! info
    "Warn" and "hide" statuses policy values only apply to non-followed accounts. This means that if you follow an account from a domain whose posts would otherwise be filtered in this way, the filter will not be applied and the posts will still appear in all the usual places you would expect to see them.

    As such, you can use this policy to functionally hide all posts from a limited domain *except* for the posts of those accounts that are worth following.

## Accounts Policy

You can apply an accounts policy to mute (aka silence) accounts from the limited domain by default. If you set this policy value to "mute", then posts by muted accounts will not appear in most timelines or threads. This is similar to how the statuses policy "hide" option works.

!!! info
    As with statuses policy, this policy only applies to non-followed accounts. For example, if user A from this instance follows user B from the limited domain, user B will not be muted from user A's perspective. However if user A from this instance does *not* follow user B from the limited domain, user B will be muted from user A's perspective.
