from playwright.sync_api import sync_playwright, expect
import time

def verify_activity_monitor():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        # Setup context with localStorage for auth
        context = browser.new_context()

        # Inject auth token (mock)
        page = context.new_page()

        # We need to hit the page first to set localstorage usually, but for this app
        # we can just set it if we navigate to a page that doesn't redirect immediately or
        # we can use add_init_script to pre-populate localStorage

        page.add_init_script("""
            localStorage.setItem('auth_token', 'mock-token');
            localStorage.setItem('user', JSON.stringify({
                username: 'admin',
                isAdmin: true,
                accountType: 'system'
            }));
        """)

        try:
            # Navigate to desktop
            page.goto("http://localhost:5173/desktop")

            # Wait for desktop to load
            expect(page.get_by_text("Mynt NAS")).to_be_visible(timeout=10000)

            # Find Activity Monitor icon and click it
            # It's in the grid, named "Activity Monitor"
            activity_icon = page.get_by_role("button", name="Activity Monitor").first
            expect(activity_icon).to_be_visible()
            activity_icon.click()

            # Wait for window to open
            expect(page.get_by_text("Activity Monitor").nth(1)).to_be_visible() # Title in window bar

            # Wait for content
            # Since backend isn't running on 8080 (or we didn't start it), API calls will fail
            # But we should see the UI structure (Tabs, Table headers)
            expect(page.get_by_role("button", name="CPU")).to_be_visible()
            expect(page.get_by_role("button", name="Memory")).to_be_visible()

            # Check table headers
            expect(page.get_by_text("Process Name")).to_be_visible()
            expect(page.get_by_text("% CPU")).to_be_visible()

            # Take screenshot
            time.sleep(1) # Wait for animations
            page.screenshot(path="verification/activity_monitor.png")
            print("Screenshot captured")

        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    verify_activity_monitor()
