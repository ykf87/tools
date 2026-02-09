package main

//	func init() {
//		s := scheduler.New(mainsignal.MainCtx)
//		rt := s.NewRunner(func(ctx context.Context) error {
//			fmt.Println("执行测试代码----")
//			return fmt.Errorf("错误咯---")
//		}, time.Second*5, nil)
//		rt.Every(time.Second * 1).RunNow()
//	}
func init() {
	// tk, err := task.NewTask("test", 1, "测试task", 5, true)
	// if err != nil {
	// 	fmt.Println("构建任务失败", err)
	// 	return
	// }
	// tr, err := tk.AddChild("text-1", "测试执行", time.Minute*60)
	// if err != nil {
	// 	fmt.Println("构建子任务失败", err)
	// 	return
	// }
	// tr.StartAtTime(-28800000, func(tr *task.TaskRun) error {
	// 	fmt.Println("-------执行", tr.RunID, tr.Title)

	// 	if tr.GetTried() >= 1 {
	// 		return nil
	// 	}
	// 	return fmt.Errorf("错误的任务:%s", tr.RunID)
	// })
}
