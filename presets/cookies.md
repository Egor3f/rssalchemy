# Capturing cookies

1. Open incognito window (don't use existing session, otherwise you will be kicked from account)
2. Log in to website you want to convert (e.g. youtube)
3. Be sure that desired page opens successfully (e.g. you can see your youtube subscriptions/posts/...)
4. Open devtools (F12). Go to network tab
5. Refresh page
6. Open the very first entry
7. In "request headers" find `Cookie:`. Copy the entire value to your RSS reader
   - Miniflux: Feed settings page, field `Set Cookies`
   - Other RSS readers are not tested yet

That's all! :)
