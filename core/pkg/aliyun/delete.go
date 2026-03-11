/**
 * @Author:         yi
 * @Description:    delete
 * @Version:        1.0.0
 * @Date:           2022/12/9 11:46
 */
package aliyun

// 删除对象
func (this *AliClient) Remove(filepath string) error {
	// 初始化oss 客户端
	client, err := this.OssClient()
	if err != nil {
		return err
	}

	// bucket
	bucket, err := client.Bucket(this.Conf.Bucket)
	if err != nil {
		return err
	}

	// 删除对象
	err = bucket.DeleteObject(filepath)
	if err != nil {
		return err
	}

	return nil
}

// 删除对象
func (this *AliClient) BulkDelete(files []string) (int, error) {
	// 初始化oss 客户端
	client, err := this.OssClient()
	if err != nil {
		return 0, err
	}

	// bucket
	bucket, err := client.Bucket(this.Conf.Bucket)
	if err != nil {
		return 0, err
	}

	// 删除对象
	res, err := bucket.DeleteObjects(files)
	if err != nil {
		return 0, err
	}

	return len(res.DeletedObjects), nil
}
