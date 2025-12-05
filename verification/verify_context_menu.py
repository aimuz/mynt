from playwright.sync_api import sync_playwright

def verify_context_menu():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        # Create a context with viewport similar to a desktop
        context = browser.new_context(viewport={"width": 1280, "height": 800})
        page = context.new_page()

        # Inject auth token into localStorage to bypass login
        # We need to navigate to the origin first to set localStorage
        page.goto("http://localhost:5173/")
        page.evaluate("""() => {
            localStorage.setItem('auth_token', 'mock-token');
            localStorage.setItem('user', JSON.stringify({username: 'admin', is_admin: true}));
        }""")

        # Navigate to the desktop page
        try:
            page.goto("http://localhost:5173/desktop", timeout=10000)

            # Wait for desktop to load (wait for menubar or something specific)
            page.wait_for_selector(".desktop-menubar", timeout=10000)
            print("Desktop loaded successfully.")

            # 1. Right click on desktop background (should show context menu)
            # Find an empty spot. The root div has the background.
            # We can click at coordinates (e.g. 500, 500) which should be empty
            page.mouse.move(500, 500)
            page.mouse.down(button="right")
            page.mouse.up(button="right")

            # Check if context menu is visible
            # The context menu has text "Change Wallpaper"
            try:
                page.wait_for_selector("text=Change Wallpaper", timeout=2000)
                print("SUCCESS: Context menu appeared on desktop background.")
            except:
                print("FAILURE: Context menu did NOT appear on desktop background.")

            # Take a screenshot
            page.screenshot(path="verification/desktop_context_menu.png")

            # Close the context menu by clicking elsewhere
            page.mouse.click(500, 500)
            page.wait_for_timeout(500) # Wait for it to close

            # 2. Right click on Menu Bar (should NOT show context menu)
            menubar = page.locator(".desktop-menubar")
            menubar.click(button="right")

            try:
                if page.locator("text=Change Wallpaper").is_visible():
                    print("FAILURE: Context menu appeared on Menu Bar.")
                else:
                    print("SUCCESS: Context menu did NOT appear on Menu Bar.")
            except:
                print("SUCCESS: Context menu did NOT appear on Menu Bar (exception).")

            # 3. Right click on an App Icon (should NOT show context menu)
            # Find the first button in the app grid
            # .flex-1.grid is the app grid container. buttons are inside.
            app_icon = page.locator(".flex-1.grid button").first
            if app_icon.count() > 0:
                app_icon.click(button="right")
                try:
                    if page.locator("text=Change Wallpaper").is_visible():
                        print("FAILURE: Context menu appeared on App Icon.")
                    else:
                        print("SUCCESS: Context menu did NOT appear on App Icon.")
                except:
                    print("SUCCESS: Context menu did NOT appear on App Icon (exception).")
            else:
                print("WARNING: Could not find app icon to test.")

            # 4. Right click on Dock (should NOT show context menu)
            dock = page.locator(".desktop-dock")
            if dock.count() > 0:
                dock.click(button="right")
                try:
                    if page.locator("text=Change Wallpaper").is_visible():
                        print("FAILURE: Context menu appeared on Dock.")
                    else:
                        print("SUCCESS: Context menu did NOT appear on Dock.")
                except:
                    print("SUCCESS: Context menu did NOT appear on Dock (exception).")
            else:
                print("WARNING: Could not find dock to test.")

            # Take another screenshot to show no context menu
            page.screenshot(path="verification/no_context_menu.png")

        except Exception as e:
            print(f"Error during verification: {e}")
            # Take screenshot of error state
            page.screenshot(path="verification/error.png")

        finally:
            browser.close()

if __name__ == "__main__":
    verify_context_menu()
