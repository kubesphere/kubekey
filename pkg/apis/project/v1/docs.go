/*
Copyright 2023 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

// Playbook keyword in ansible: https://docs.ansible.com/ansible/latest/reference_appendices/playbooks_keywords.html#playbook-keywords
// support list (base on ansible 2.15.5)

/**
Play
+------+------------------------+------------+
| Row  |        Keyword         |  Support   |
+------+------------------------+------------+
|   1  |   any_errors_fatal     |     ✘      |
|   2  |   become               |     ✘      |
|   3  |   become_exe           |     ✘      |
|   4  |   become_flags         |     ✘      |
|   5  |   become_method        |     ✘      |
|   6  |   become_user          |     ✘      |
|   7  |   check_mode           |     ✘      |
|   8  |   collections          |     ✘      |
|   9  |   connection           |     ✔︎      |
|  10  |   debugger             |     ✘      |
|  11  |   diff                 |     ✘      |
|  12  |   environment          |     ✘      |
|  13  |   fact_path            |     ✘      |
|  14  |   force_handlers       |     ✘      |
|  15  |   gather_facts         |     ✔︎      |
|  16  |   gather_subset        |     ✘      |
|  17  |   gather_timeout       |     ✘      |
|  18  |   handlers             |     ✘      |
|  19  |   hosts                |     ✔︎      |
|  20  |   ignore_errors        |     ✔︎      |
|  21  |   ignore_unreachable   |     ✘      |
|  22  |   max_fail_percentage  |     ✘      |
|  23  |   module_defaults      |     ✘      |
|  24  |   name                 |     ✔︎      |
|  25  |   no_log               |     ✘      |
|  26  |   order                |     ✘      |
|  27  |   port                 |     ✘      |
|  28  |   post_task            |     ✔︎      |
|  29  |   pre_tasks            |     ✔︎      |
|  30  |   remote_user          |     ✘      |
|  31  |   roles                |     ✔︎      |
|  32  |   run_once             |     ✔︎      |
|  33  |   serial               |     ✔︎      |
|  34  |   strategy             |     ✘      |
|  35  |   tags                 |     ✔︎      |
|  36  |   tasks                |     ✔︎      |
|  37  |   throttle             |     ✘      |
|  38  |   timeout              |     ✘      |
|  39  |   vars                 |     ✔︎      |
|  40  |   vars_files           |     ✘      |
|  41  |   vars_prompt          |     ✘      |
+------+------------------------+------------+

Role
+------+------------------------+------------+
| Row  |        Keyword         |  Support   |
+------+------------------------+------------+
|   1  |   any_errors_fatal     |     ✘      |
|   2  |   become               |     ✘      |
|   3  |   become_exe           |     ✘      |
|   4  |   become_flags         |     ✘      |
|   5  |   become_method        |     ✘      |
|   6  |   become_user          |     ✘      |
|   7  |   check_mode           |     ✘      |
|   8  |   collections          |     ✘      |
|   9  |   connection           |     ✘      |
|  10  |   debugger             |     ✘      |
|  11  |   delegate_facts       |     ✘      |
|  12  |   delegate_to          |     ✘      |
|  13  |   diff                 |     ✘      |
|  14  |   environment          |     ✘      |
|  15  |   ignore_errors        |     ✔︎      |
|  16  |   ignore_unreachable   |     ✘      |
|  17  |   max_fail_percentage  |     ✘      |
|  18  |   module_defaults      |     ✘      |
|  19  |   name                 |     ✔︎      |
|  20  |   no_log               |     ✘      |
|  21  |   port                 |     ✘      |
|  22  |   remote_user          |     ✘      |
|  23  |   run_once             |     ✔︎      |
|  24  |   tags                 |     ✔︎      |
|  25  |   throttle             |     ✘      |
|  26  |   timeout              |     ✘      |
|  27  |   vars                 |     ✔︎      |
|  28  |   when                 |     ✔︎      |
+------+------------------------+------------+

Block
+------+------------------------+------------+
| Row  |        Keyword         |  Support   |
+------+------------------------+------------+
|   1  |   always               |     ✔︎      |
|   2  |   any_errors_fatal     |     ✘      |
|   3  |   become               |     ✘      |
|   4  |   become_exe           |     ✘      |
|   5  |   become_flags         |     ✘      |
|   6  |   become_method        |     ✘      |
|   7  |   become_user          |     ✘      |
|   8  |   block                |     ✔︎      |
|   9  |   check_mode           |     ✘      |
|  10  |   collections          |     ✘      |
|  11  |   debugger             |     ✘      |
|  12  |   delegate_facts       |     ✘      |
|  13  |   delegate_to          |     ✘      |
|  14  |   diff                 |     ✘      |
|  15  |   environment          |     ✘      |
|  16  |   ignore_errors        |     ✔︎      |
|  17  |   ignore_unreachable   |     ✘      |
|  18  |   max_fail_percentage  |     ✘      |
|  19  |   module_defaults      |     ✘      |
|  20  |   name                 |     ✔︎      |
|  21  |   no_log               |     ✘      |
|  22  |   notify               |     ✘      |
|  23  |   port                 |     ✘      |
|  24  |   remote_user          |     ✘      |
|  25  |   rescue               |     ✔︎      |
|  26  |   run_once             |     ✘      |
|  27  |   tags                 |     ✔︎      |
|  28  |   throttle             |     ✘      |
|  29  |   timeout              |     ✘      |
|  30  |   vars                 |     ✔︎      |
|  31  |   when                 |     ✔︎      |
+------+------------------------+------------+


Task
+------+------------------------+------------+
| Row  |        Keyword         |  Support   |
+------+------------------------+------------+
|   1  |   action               |     ✔︎      |
|   2  |   any_errors_fatal     |     ✘      |
|   3  |   args                 |     ✔︎      |
|   4  |   async                |     ✘      |
|   5  |   become               |     ✘      |
|   6  |   become_exe           |     ✘      |
|   7  |   become_flags         |     ✘      |
|   8  |   become_method        |     ✘      |
|   9  |   become_user          |     ✘      |
|  10  |   changed_when         |     ✘      |
|  11  |   check_mode           |     ✘      |
|  12  |   collections          |     ✘      |
|  13  |   debugger             |     ✘      |
|  14  |   delay                |     ✘      |
|  15  |   delegate_facts       |     ✘      |
|  16  |   delegate_to          |     ✘      |
|  17  |   diff                 |     ✘      |
|  18  |   environment          |     ✘      |
|  19  |   failed_when          |     ✔︎      |
|  20  |   ignore_errors        |     ✔︎      |
|  21  |   ignore_unreachable   |     ✘      |
|  22  |   local_action         |     ✘      |
|  23  |   loop                 |     ✔︎      |
|  24  |   loop_control         |     ✘      |
|  25  |   module_defaults      |     ✘      |
|  26  |   name                 |     ✔︎      |
|  27  |   no_log               |     ✘      |
|  28  |   notify               |     ✘      |
|  29  |   poll                 |     ✘      |
|  30  |   port                 |     ✘      |
|  31  |   register             |     ✔︎      |
|  32  |   remote_user          |     ✘      |
|  33  |   retries              |     ✘      |
|  34  |   run_once             |     ✘      |
|  35  |   tags                 |     ✔︎      |
|  36  |   throttle             |     ✘      |
|  37  |   timeout              |     ✘      |
|  38  |   until                |     ✘      |
|  39  |   vars                 |     ✔︎      |
|  40  |   when                 |     ✔︎      |
|  41  |   with_<lookup_plugin> |     ✔︎      |
+------+------------------------+------------+
*/
