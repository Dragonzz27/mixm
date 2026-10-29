package app

import (
	"context"
	"time"

	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
)

type exampleLesson struct {
	key        string
	name       string
	title      string
	kind       string
	body       string
	assignment string
}

type exampleQuestion struct {
	key         string
	lesson      string
	title       string
	kind        string
	answer      string
	explanation string
	options     []model.Record
}

var exampleLessons = []exampleLesson{
	{
		key:   "linux",
		name:  "Linux 基础",
		title: "【示例】Linux 终端与文件操作",
		kind:  "tutorial",
		body: `[md]# Linux 终端与文件操作

本内容由 initdb 创建，用于体验内容浏览和作业提交流程，可以编辑或删除。

## 学习目标

- 使用 pwd 查看当前工作目录。
- 使用 ls 查看目录中的文件。
- 使用 mkdir 创建目录，使用 cat 查看文本内容。

## 动手练习

在自己的开发终端中依次执行：

~~~bash
pwd
mkdir -p /workspace/mixm-demo
cd /workspace/mixm-demo
printf 'Hello mixm\n' > hello.txt
ls
cat hello.txt
~~~

完成后，到“我的作业”提交执行步骤和输出。`,
		assignment: "在自己的开发终端完成教学内容中的文件操作。提交使用过的命令、hello.txt 的内容，以及 pwd 和 ls 的输出。可以直接填写文字，无需上传附件。",
	},
	{
		key:   "python",
		name:  "Python 入门",
		title: "【示例】Python 列表与基础计算",
		kind:  "experiment",
		body: `[md]# Python 列表与基础计算

## 学习目标

理解列表、循环和函数的基本用法。

## 实验步骤

新建 demo.py，输入以下代码并在安装了 Python 的终端中执行 python3 demo.py：

~~~python
scores = [72, 85, 91, 68, 94]
average = sum(scores) / len(scores)
passed = [score for score in scores if score >= 80]
print("平均分:", average)
print("达到 80 分的成绩:", passed)
~~~

预期平均分为 82.0，达到 80 分的成绩为 [85, 91, 94]。

尝试修改列表中的成绩，再观察结果。`,
		assignment: "运行教学内容中的 Python 示例，提交代码和输出。再将成绩列表改为 [60, 70, 80, 90, 100]，说明平均分和达到 80 分的成绩分别是什么。",
	},
	{
		key:   "containers",
		name:  "容器与实例",
		title: "【示例】镜像、实例与持久化工作目录",
		kind:  "documentation",
		body: `[md]# 镜像、实例与持久化工作目录

## 镜像与实例

镜像保存预装的软件和启动配置；实例是按镜像及资源配置创建的运行环境。同一镜像可以用于创建多个不同资源规格的实例。

## 资源配置

- CPU 的单位是核，内存可使用 Mi 或 Gi。
- GPU 数量和单卡显存分别配置。
- 临时存储限制容器的临时空间。

## 工作目录

项目将 /workspace 挂载到用户的持久化工作卷。临时存储额度和持久化工作卷容量由不同配置控制。

## 思考

为什么同一个镜像可以创建多个实例？为什么需要将需要保留的文件放在持久化工作目录中？`,
		assignment: "用自己的话说明镜像和实例的关系，并解释容器临时存储与 /workspace 持久化工作卷的区别。给出一个需要保存到持久化目录中的文件示例。",
	},
}

var exampleQuestions = []exampleQuestion{
	{
		key:         "linux-pwd",
		lesson:      "linux",
		title:       "【示例·单选】哪个命令用于查看当前工作目录？",
		kind:        "single_choice",
		answer:      "A",
		explanation: "pwd 输出当前工作目录；ls 列出目录内容，mkdir 创建目录，cat 查看文件。",
		options: []model.Record{
			{
				"option_key": "A",
				"content":    "pwd",
				"is_correct": true,
			},
			{
				"option_key": "B",
				"content":    "ls",
				"is_correct": false,
			},
			{
				"option_key": "C",
				"content":    "mkdir",
				"is_correct": false,
			},
			{
				"option_key": "D",
				"content":    "cat",
				"is_correct": false,
			},
		},
	},
	{
		key:         "python-containers",
		lesson:      "python",
		title:       "【示例·多选】哪些 Python 类型可以保存多个元素？",
		kind:        "multi_choice",
		answer:      "A,B",
		explanation: "列表和字典是容器类型；int 和 bool 分别表示整数与布尔值。",
		options: []model.Record{
			{
				"option_key": "A",
				"content":    "list（列表）",
				"is_correct": true,
			},
			{
				"option_key": "B",
				"content":    "dict（字典）",
				"is_correct": true,
			},
			{
				"option_key": "C",
				"content":    "int（整数）",
				"is_correct": false,
			},
			{
				"option_key": "D",
				"content":    "bool（布尔值）",
				"is_correct": false,
			},
		},
	},
	{
		key:         "persistent-workspace",
		lesson:      "containers",
		title:       "【示例·判断】容器临时存储额度就是 /workspace 持久化工作卷的容量。",
		kind:        "true_false",
		answer:      "false",
		explanation: "临时存储与持久化工作卷分别配置，storage_limit 不改变工作卷容量。",
	},
	{
		key:         "python-print",
		lesson:      "python",
		title:       "【示例·填空】Python 中用于输出文本的内置函数名是 ____。",
		kind:        "blank",
		answer:      "print",
		explanation: "例如 print('Hello mixm')，填写函数名 print 即可。",
	},
	{
		key:         "image-instance",
		lesson:      "containers",
		title:       "【示例·简答】简述镜像与实例的关系。",
		kind:        "essay",
		answer:      "镜像保存运行环境，实例是基于镜像创建的运行环境。",
		explanation: "镜像可复用，同一个镜像可以创建多个实例。本题用于演示简答题填写，当前自动判分按参考答案文本匹配。",
	},
}

// seedExamples 组织开发示例与分类、选项及练习关联，不依赖管理员创建顺序。
func seedExamples(ctx context.Context, db *repository.DB) error {
	rootID, _, err := repository.SeedExample(ctx, db, "categories", "root", model.Record{
		"name":        "【示例】开发测试课程",
		"description": "initdb 创建的教学、作业和题目示例，可编辑或删除。",
		"level":       1,
		"sort_order":  0,
		"is_active":   true,
	})
	if err != nil {
		return err
	}

	categoryIDs := map[string]string{}
	for index, lesson := range exampleLessons {
		categoryID, _, err := repository.SeedExample(ctx, db, "categories", lesson.key, model.Record{
			"name":       lesson.name,
			"parent_id":  rootID,
			"level":      2,
			"sort_order": index,
			"is_active":  true,
		})
		if err != nil {
			return err
		}
		categoryIDs[lesson.key] = categoryID

		contentID, created, err := repository.SeedExample(ctx, db, "contents", lesson.key, model.Record{
			"title":        lesson.title,
			"content":      lesson.body,
			"content_type": lesson.kind,
			"status":       "published",
			"is_active":    true,
		})
		if err != nil {
			return err
		}
		if created {
			if _, _, err = repository.SeedExample(ctx, db, "content_categories", lesson.key, model.Record{
				"content_id":  contentID,
				"category_id": categoryID,
				"sort_order":  index,
			}); err != nil {
				return err
			}
		}

		// 使用系统来源标记，避免初始化额外创建可登录账号；截止时间只在首次插入时设置。
		if _, _, err = repository.SeedExample(ctx, db, "assignments", lesson.key, model.Record{
			"title":       "【示例作业】" + lesson.name,
			"description": lesson.assignment,
			"attachments": "[]",
			"content_id":  contentID,
			"deadline":    time.Now().AddDate(0, 1, 0),
			"status":      "active",
			"created_by":  "initdb",
		}); err != nil {
			return err
		}
	}

	// 题目首次创建时补充选项与知识分类，保留管理员后续修改过的关联。
	questionIDs := []string{}
	for _, question := range exampleQuestions {
		id, created, err := repository.SeedExample(ctx, db, "questions", question.key, model.Record{
			"title":           question.title,
			"type":            question.kind,
			"content":         question.title,
			"answer":          question.answer,
			"explanation":     question.explanation,
			"difficulty":      1,
			"tags":            model.JSON([]string{"开发测试示例"}),
			"source":          "initdb 开发示例",
			"created_by":      "initdb",
			"is_ai_generated": false,
			"review_status":   "approved",
		})
		if err != nil {
			return err
		}
		questionIDs = append(questionIDs, id)
		if !created {
			continue
		}

		for _, option := range question.options {
			data := model.Record{"question_id": id}
			for key, value := range option {
				data[key] = value
			}
			if _, _, err = repository.SeedExample(ctx, db, "question_options", question.key+":"+model.String(option, "option_key"), data); err != nil {
				return err
			}
		}
		if _, err := repository.InsertResourceLink(ctx, db, "category_id", "question", "question_tags", id, categoryIDs[question.lesson]); err != nil {
			return err
		}
	}

	// 将五种题型关联到一份练习，便于直接验证组卷与答题流程。
	practiceID, created, err := repository.SeedExample(ctx, db, "practices", "intro", model.Record{
		"title":       "【示例练习】Linux、Python 与容器基础",
		"description": "包含单选、多选、判断、填空和简答五种题型，用于开发测试。",
		"created_by":  "initdb",
		"type_counts": model.JSON(model.Record{
			"single_choice": 1,
			"multi_choice":  1,
			"true_false":    1,
			"fill_blank":    1,
			"short_answer":  1,
		}),
	})
	if err != nil || !created {
		return err
	}
	for index, id := range questionIDs {
		if _, _, err = repository.SeedExample(ctx, db, "practice_questions", exampleQuestions[index].key, model.Record{
			"practice_id": practiceID,
			"question_id": id,
			"sort_order":  index,
		}); err != nil {
			return err
		}
	}
	return nil
}
