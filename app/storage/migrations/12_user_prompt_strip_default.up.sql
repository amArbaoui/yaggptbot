update user_prompt
set prompt = substr(prompt, instr(prompt, 'USER_PROMPT: ') + length('USER_PROMPT: '))
where instr(prompt, 'USER_PROMPT: ') > 0;
