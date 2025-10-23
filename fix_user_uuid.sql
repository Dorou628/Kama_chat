-- 查找老柳ovo用户信息
SELECT * FROM user_info WHERE nickname LIKE '%老柳%' OR nickname LIKE '%ovo%';

-- 查找所有用户的UUID格式，了解现有的UUID模式
SELECT uuid, nickname, telephone, created_at FROM user_info ORDER BY created_at;

-- 如果找到了老柳ovo用户，可以选择以下操作之一：

-- 选项1：修改老柳ovo的UUID为标准格式（假设找到的UUID是某个值）
-- UPDATE user_info SET uuid = 'U00000000999' WHERE nickname LIKE '%老柳%' OR nickname LIKE '%ovo%';

-- 选项2：软删除老柳ovo账号（推荐，保留数据）
-- UPDATE user_info SET deleted_at = NOW() WHERE nickname LIKE '%老柳%' OR nickname LIKE '%ovo%';

-- 选项3：完全删除老柳ovo账号（不推荐，会丢失数据）
-- DELETE FROM user_info WHERE nickname LIKE '%老柳%' OR nickname LIKE '%ovo%';

-- 查看最大的UUID编号，为新的UUID分配做准备
SELECT MAX(CAST(SUBSTRING(uuid, 2) AS UNSIGNED)) as max_uuid_number FROM user_info WHERE uuid REGEXP '^U[0-9]+$';