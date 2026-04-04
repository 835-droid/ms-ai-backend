// MongoDB script to fix user roles and check authentication issues
// Run this in mongosh or mongo shell

// 1. First, check current users and their roles
print("\n=== Current Users ===");
db.users.find({}, {
    _id: 1,
    username: 1,
    user_id: 1,
    roles: 1,
    is_active: 1,
    created_at: 1
}).forEach(user => {
    print(`User: ${user.username} (ID: ${user._id})`);
    print(`  Roles: ${JSON.stringify(user.roles)}`);
    print(`  Active: ${user.is_active}`);
    print(`  Public ID: ${user.user_id}`);
    print("");
});

// 2. Check if any user has admin role
const adminCount = db.users.countDocuments({ roles: "admin" });
print(`\n=== Admin Users Count: ${adminCount} ===`);

if (adminCount === 0) {
    print("⚠️  WARNING: No admin users found!");
    print("You need to promote a user to admin to access admin endpoints.");
}

// 3. Function to promote a user to admin
function promoteToAdmin(username) {
    const result = db.users.updateOne(
        { username: username },
        { 
            $addToSet: { roles: "admin" },
            $set: { is_active: true, updated_at: new Date() }
        }
    );
    
    if (result.modifiedCount > 0) {
        print(`✅ Successfully promoted user '${username}' to admin`);
        const user = db.users.findOne({ username: username });
        print(`   New roles: ${JSON.stringify(user.roles)}`);
    } else if (result.matchedCount === 0) {
        print(`❌ User '${username}' not found`);
    } else {
        print(`ℹ️  User '${username}' already has admin role`);
    }
}

// 4. Function to check user's JWT claims
function checkUserToken(username) {
    const user = db.users.findOne({ username: username });
    if (!user) {
        print(`❌ User '${username}' not found`);
        return;
    }
    
    print(`\n=== User '${username}' Details ===`);
    print(`MongoDB ID: ${user._id}`);
    print(`Public ID: ${user.user_id}`);
    print(`Roles: ${JSON.stringify(user.roles)}`);
    print(`Active: ${user.is_active}`);
    print(`Has refresh token: ${!!user.refresh_token}`);
    
    // Check if user can access admin endpoints
    const hasAdmin = user.roles && user.roles.includes("admin");
    print(`\nCan access admin endpoints: ${hasAdmin ? '✅ YES' : '❌ NO'}`);
    
    if (!hasAdmin) {
        print("\nTo fix this, run: promoteToAdmin('" + username + "')");
    }
}

// 5. Function to list all users with their permissions
function listUserPermissions() {
    print("\n=== User Permissions ===");
    db.users.find({}, {
        username: 1,
        roles: 1,
        is_active: 1
    }).forEach(user => {
        const canAdmin = user.roles && user.roles.includes("admin");
        const status = user.is_active ? "✅ Active" : "❌ Inactive";
        print(`${user.username.padEnd(20)} | Admin: ${canAdmin ? '✅' : '❌'} | Status: ${status}`);
    });
}

// Usage examples:
print("\n=== Usage Examples ===");
print("promoteToAdmin('username')     - Promote a user to admin");
print("checkUserToken('username')     - Check user's authentication details");
print("listUserPermissions()          - List all users and their permissions");

// Auto-run diagnostics if a username is provided as argument
if (typeof username !== 'undefined') {
    checkUserToken(username);
}