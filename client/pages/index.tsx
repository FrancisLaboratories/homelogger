"use client";

import React, { useEffect, useState } from 'react';
import { Container, Form } from 'react-bootstrap';
import "bootstrap-icons/font/bootstrap-icons.css";
import MyNavbar from '../components/Navbar';
import ListGroup from 'react-bootstrap/ListGroup';
import TodoItem from '../components/TodoItem';
import { SERVER_URL } from "@/pages/_app";

const todoUrl = `${SERVER_URL}/todo`;
const todoAddUrl = `${SERVER_URL}/todo/add`;

type Todo = {
  id: string | number;
  label: string;
  checked: boolean;
  applianceId?: number | null;
  spaceType?: string | null;
  sourceLabel?: string | null;
  createdAt?: string | null;
  CreatedAt?: string | null;
  created_at?: string | null;
};

const HomePage: React.FC = () => {
  const [todos, setTodos] = useState<Todo[]>([]);
  const [sortOption, setSortOption] = useState<string>('created_desc');
  const [filterOption, setFilterOption] = useState<string>('all');

  useEffect(() => {
    const fetchTodos = async () => {
      try {
        const response = await fetch(todoUrl);
        const data = await response.json();

        const dataTyped: Todo[] = data as Todo[];
        const applianceIds: number[] = Array.from(new Set(dataTyped.filter((t) => t.applianceId).map((t) => Number(t.applianceId))));
        const nameMap: Record<number, string> = {};
        await Promise.all(applianceIds.map(async (id) => {
          try {
            const r = await fetch(`${SERVER_URL}/appliances/${id}`);
            if (!r.ok) return;
            const a = await r.json();
            nameMap[id] = a.applianceName || `Appliance ${id}`;
          } catch (e) {
            console.error('Error loading appliance name', e);
          }
        }));

        const enriched: Todo[] = dataTyped.map((t) => ({
          ...t,
          sourceLabel: t.applianceId ? nameMap[Number(t.applianceId)] : (t.spaceType || null),
        }));

        setTodos(enriched);
      } catch (error) {
        console.error('Error fetching todos:', error);
      }
    };

    fetchTodos();
  }, []);

  const handleAddTodo = async () => {
    const label = prompt('What should go in this item?');
    if (label) {
      const newTodo = { label, checked: false, userid: "1" };

      try {
        const response = await fetch(todoAddUrl, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(newTodo),
        });

        if (!response.ok) {
          throw new Error('Failed to add todo');
        }

        const addedTodo = await response.json();
        setTodos([...todos, addedTodo]);
      } catch (error) {
        console.error('Error adding todo:', error);
      }
    }
  };

  const handleDeleteTodo = (id: string) => {
    setTodos((prevTodos) => prevTodos.filter((todo) => String(todo.id) !== id));
  };

  return (
    <Container>
      <MyNavbar />
      <h4 id='maintext'>To-dos:</h4>

      <div style={{display: 'flex', gap: '8px', marginBottom: '8px'}}>
        <Form.Select aria-label="Sort todos" value={sortOption} onChange={(e) => setSortOption(e.target.value)} style={{maxWidth: '220px'}}>
          <option value="created_desc">Created (newest)</option>
          <option value="created_asc">Created (oldest)</option>
          <option value="label_asc">Label (A - Z)</option>
          <option value="label_desc">Label (Z - A)</option>
        </Form.Select>

        <Form.Select aria-label="Filter todos" value={filterOption} onChange={(e) => setFilterOption(e.target.value)} style={{maxWidth: '180px'}}>
          <option value="all">All</option>
          <option value="completed">Completed</option>
          <option value="not_completed">Not completed</option>
        </Form.Select>
      </div>

      <ListGroup>
        {(() => {
          const filtered = todos.filter((t) => {
            if (filterOption === 'all') return true;
            if (filterOption === 'completed') return !!t.checked;
            if (filterOption === 'not_completed') return !t.checked;
            return true;
          });

          const sorted = filtered.slice().sort((a, b) => {
            const sa = (a.label || '').toString();
            const sb = (b.label || '').toString();

            const ca = a.createdAt || a.CreatedAt || a.created_at || null;
            const cb = b.createdAt || b.CreatedAt || b.created_at || null;

            if (sortOption === 'label_asc') return sa.localeCompare(sb);
            if (sortOption === 'label_desc') return sb.localeCompare(sa);

            const da = ca ? new Date(ca).getTime() : 0;
            const db = cb ? new Date(cb).getTime() : 0;
            if (sortOption === 'created_asc') return da - db || sa.localeCompare(sb);
            return db - da || sa.localeCompare(sb);
          });

          return sorted.map((todo, index) => (
            <TodoItem key={index} id={String(todo.id)} label={todo.label} checked={todo.checked} onDelete={handleDeleteTodo} applianceId={todo.applianceId || undefined} spaceType={todo.spaceType || undefined} sourceLabel={todo.sourceLabel || undefined} createdAt={todo.createdAt || todo.CreatedAt || todo.created_at || undefined} />
          ));
        })()}
      </ListGroup>
      <i className="bi bi-plus-square-fill" onClick={handleAddTodo} style={{ fontSize: '2rem', cursor: "pointer" }}></i>
    </Container>
  );
};

export default HomePage;
