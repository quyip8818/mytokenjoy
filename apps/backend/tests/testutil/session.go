package testutil

const SessionCookieAdmin = "tokenjoy_session_member=m-admin"

func SessionCookie(memberID string) string {
	return "tokenjoy_session_member=" + memberID
}
