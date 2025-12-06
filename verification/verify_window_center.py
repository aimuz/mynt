import json
from playwright.sync_api import Page, expect, sync_playwright

def verify_window_centering(page: Page):
    # 1. Bypass authentication
    page.goto("http://localhost:5173/login")

    # Inject auth token and user into localStorage
    auth_data = {
        "token": "fake-token-for-testing",
        "user": {"username": "admin", "role": "admin"}
    }

    page.evaluate(f"""
        localStorage.setItem('auth_token', '{auth_data["token"]}');
        localStorage.setItem('user', '{json.dumps(auth_data["user"])}');
    """)

    # 2. Go to desktop
    page.goto("http://localhost:5173/desktop")

    # Wait for desktop to load (wait for an icon)
    page.wait_for_selector(".desktop-icon")

    # 3. Open an app (e.g., Settings)
    # Using a more specific selector to avoid ambiguity
    # This targets the button that contains the text "Settings" and has the class "desktop-icon"
    page.locator("button.desktop-icon").filter(has_text="Settings").first.click()

    # 4. Wait for window to appear
    window_selector = ".desktop-window"
    page.wait_for_selector(window_selector)

    # 5. Take screenshot
    page.screenshot(path="verification/verification.png")

    # 6. Verify coordinates programmatically
    window_element = page.locator(window_selector).first
    box = window_element.bounding_box()
    viewport = page.viewport_size

    if box and viewport:
        print(f"Window Box: {box}")
        print(f"Viewport: {viewport}")

        # Expected center (800x600 window)
        expected_x = (viewport['width'] - 800) / 2
        expected_y = (viewport['height'] - 600) / 2

        print(f"Expected X ~ {expected_x}, Y ~ {expected_y}")

        # Check if within 5px (to account for potentially minor rendering diffs, though math should be exact)
        if abs(box['x'] - expected_x) < 5 and abs(box['y'] - expected_y) < 5:
            print("VERIFICATION SUCCESS: Window is centered.")
        else:
            print("VERIFICATION FAILURE: Window is not centered.")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        # Set a fixed viewport size to make calculation deterministic
        page = browser.new_page(viewport={"width": 1280, "height": 800})
        try:
            verify_window_centering(page)
        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()
