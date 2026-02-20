
import requests
import json
import random
import string
import time

# 配置
USER_SERVICE_URL = "https://we-circle.zeabur.app"
POST_SERVICE_URL = "https://we-circle-post.zeabur.app"

# 辅助函数：格式化打印 JSON
def print_json(label, data):
    print(f"[{label}]:")
    print(json.dumps(data, indent=2, ensure_ascii=False))
    print("-" * 40)

# 辅助函数：发送请求并打印详情
def make_request(method, url, data=None, headers=None, service_name=""):
    print(f"\n🚀 Requesting {service_name}...")
    print(f"URL: {url}")
    print(f"Method: {method}")
    if headers:
        print_json("Headers", headers)
    if data:
        print_json("Payload", data)

    try:
        if method.upper() == "POST":
            resp = requests.post(url, json=data, headers=headers)
        elif method.upper() == "GET":
            resp = requests.get(url, params=data, headers=headers)
        else:
            print(f"Unsupported method: {method}")
            return None

        print(f"Status: {resp.status_code}")
        
        try:
            resp_json = resp.json()
            print_json("Response", resp_json)
            return resp, resp_json
        except ValueError:
            print(f"[Response Text]: {resp.text}")
            return resp, None
            
    except Exception as e:
        print(f"❌ Request Error: {e}")
        return None, None

# 随机生成用户数据
def generate_user():
    rand_str = ''.join(random.choices(string.ascii_lowercase + string.digits, k=8))
    username = f"user_{rand_str}"
    email = f"{username}@example.com"
    password = "Password123!"
    return username, email, password

def print_step(message):
    print(f"\n========================================")
    print(f"➜ {message}")
    print(f"========================================")

def test_workflow():
    username, email, password = generate_user()
    print(f"生成的测试用户: {username} / {email}")

    # ==========================================
    # 1. 注册用户
    # ==========================================
    print_step("1. 注册用户 (User Service)")
    register_url = f"{USER_SERVICE_URL}/api/v1/users/register"
    register_data = {
        "username": username,
        "email": email,
        "password": password
    }
    
    resp, _ = make_request("POST", register_url, register_data, service_name="User Service - Register")
    if not resp or resp.status_code != 200:
        print("❌ 注册失败")
        return
    print("✅ 注册成功")

    # ==========================================
    # 2. 登录用户
    # ==========================================
    print_step("2. 登录用户 (User Service)")
    login_url = f"{USER_SERVICE_URL}/api/v1/users/login"
    login_data = {
        "email": email,
        "password": password
    }
    
    resp, resp_json = make_request("POST", login_url, login_data, service_name="User Service - Login")
    
    token = ""
    if resp and resp.status_code == 200 and resp_json:
        token = resp_json.get("data", {}).get("accessToken")
    
    if not token:
        print("❌ 获取 Token 失败")
        return
        
    print(f"✅ 登录成功! Token: {token[:20]}...")

    # 等待一小会儿确保服务同步（如果有需要）
    time.sleep(1)

    # ==========================================
    # 3. 发布动态
    # ==========================================
    print_step("3. 发布动态 (Post Service)")
    create_post_url = f"{POST_SERVICE_URL}/api/v1/posts/create"
    post_data = {
        "title": "Hello Zeabur",
        "content": "This is a test post from automated script."
    }
    headers = {
        "Authorization": f"Bearer {token}"
    }
    
    resp, _ = make_request("POST", create_post_url, post_data, headers=headers, service_name="Post Service - Create")
    
    if not resp or resp.status_code != 200:
        print("❌ 发布动态失败")
        return
    print("✅ 发布动态成功")

    # ==========================================
    # 4. 获取所有动态列表
    # ==========================================
    print_step("4. 获取动态列表 (Post Service)")
    get_posts_url = f"{POST_SERVICE_URL}/api/v1/posts/get_all"
    query_data = {
        "page": 1,
        "pageSize": 10
    }
    
    resp, resp_json = make_request("POST", get_posts_url, query_data, headers=headers, service_name="Post Service - Get All")
    
    if not resp or resp.status_code != 200 or not resp_json:
        print("❌ 获取列表失败")
        return

    posts = resp_json.get("data", {}).get("posts", [])
    print(f"✅ 获取成功，当前共有 {len(posts)} 条动态")
    
    # 验证
    found = False
    target_content = "This is a test post from automated script."
    for post in posts:
        post_content = post.get("content", "")
        if post_content == target_content:
            found = True
            print(f"--> Found my post: ID={post.get('postId')}")
            break
    
    if found:
        print("✅ 验证成功：发布的动态存在于列表中")
    else:
        print("⚠️ 验证警告：未在列表中找到刚才发布的动态")

if __name__ == "__main__":
    print("开始 Zeabur 部署验证测试...")
    test_workflow()
