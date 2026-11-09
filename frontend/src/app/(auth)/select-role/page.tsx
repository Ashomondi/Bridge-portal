'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { 
  Bell, 
  User, 
  ArrowRight, 
  Users, 
  BookOpen, 
  FlaskConical, 
  Lock, 
  Clock, 
  HelpCircle, 
  LayoutGrid, 
  Shield, 
  Settings as SettingsIcon 
} from 'lucide-react';

interface SchoolRole {
  id: string;
  schoolName: string;
  role: 'Parent' | 'Teacher' | 'Waiting for verification';
  details: string;
  subDetails: string;
  type: 'active' | 'pending';
  badgeColor?: string;
  badgeBg?: string;
}

export default function SelectRolePage() {
  const router = useRouter();
  const [activeNav, setActiveNav] = useState('roles');

  // Sample mock roles corresponding to your design spec
  const roles: SchoolRole[] = [
    {
      id: '1',
      schoolName: 'Lakeview Secondary',
      role: 'Parent',
      details: '2 children: Amani, Zawadi',
      subDetails: 'Last used',
      type: 'active',
    },
    {
      id: '2',
      schoolName: 'Lakeview Secondary',
      role: 'Teacher',
      details: 'Class teacher • Grade 8 East',
      subDetails: 'Subjects: Mathematics',
      type: 'active',
    },
    {
      id: '3',
      schoolName: "Hillcrest Girls' High School",
      role: 'Teacher',
      details: 'Subjects: Physics, Chemistry',
      subDetails: '',
      type: 'active',
    },
    {
      id: '4',
      schoolName: 'St. Teresa Academy',
      role: 'Waiting for verification',
      details: 'Application submitted • Pending dean review',
      subDetails: '',
      type: 'pending',
    },
  ];

  const handleSelectRole = (roleItem: SchoolRole) => {
    if (roleItem.type === 'pending') return; // Cannot select pending roles
    
    // Route user to the appropriate portal dashboard based on their role
    if (roleItem.role === 'Parent') {
      router.push('/parent');
    } else if (roleItem.role === 'Teacher') {
      router.push('/teacher');
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#e0f2f1] via-[#f4f8f7] to-white flex flex-col items-center justify-between p-4 text-slate-800">
      
      {/* Top Demo Banner */}
      <div className="text-xs font-medium text-slate-600 bg-white/80 backdrop-blur px-4 py-1.5 rounded-full border border-slate-200/60 shadow-sm mt-2">
        Demo environment • sample data only
      </div>

      {/* Main Container Card */}
      <div className="w-full max-w-md bg-white/70 backdrop-blur-md border border-slate-200/90 rounded-[32px] p-6 shadow-2xl space-y-6 my-4">
        
        {/* Top Header Bar inside card */}
        <div className="flex items-center justify-between pb-3 border-b border-dashed border-slate-300">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-xl bg-[#008272] flex items-center justify-center text-white font-bold shadow-md shadow-[#008272]/20 text-sm">
              I
            </div>
            <span className="font-bold text-slate-900 tracking-tight">Role Switcher</span>
          </div>
          <div className="flex items-center gap-3">
            <button className="p-2 text-slate-600 hover:bg-slate-100 rounded-full transition">
              <Bell className="w-4 h-4" />
            </button>
            <div className="w-8 h-8 rounded-full bg-slate-900 text-white flex items-center justify-center text-xs font-semibold">
              <User className="w-4 h-4" />
            </div>
          </div>
        </div>

        {/* Title & Description */}
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-slate-900">Choose where to go</h1>
          <p className="text-slate-600 text-sm leading-relaxed">
            You have more than one role. Pick one to continue. You can switch later in your account.
          </p>
        </div>

        {/* Roles List */}
        <div className="space-y-3.5">
          {roles.map((item) => (
            <div
              key={item.id}
              onClick={() => handleSelectRole(item)}
              className={`bg-white border rounded-2xl p-4 transition shadow-sm relative group ${
                item.type === 'pending' 
                  ? 'border-slate-200/80 opacity-90 cursor-not-allowed bg-slate-50/50' 
                  : 'border-slate-200 hover:border-[#008272] hover:shadow-md cursor-pointer'
              }`}
            >
              {item.subDetails === 'Last used' && (
                <div className="flex items-center justify-between mb-1.5">
                  <span className="text-[11px] font-semibold text-purple-600 flex items-center gap-1">
                    <span className="w-1.5 h-1.5 rounded-full bg-purple-600"></span> Last used
                  </span>
                  <ArrowRight className="w-4 h-4 text-slate-400 group-hover:text-[#008272] transition" />
                </div>
              )}

              {item.type === 'active' && item.subDetails !== 'Last used' && (
                <div className="flex items-center justify-between mb-1">
                  <h2 className="font-bold text-slate-900 text-base">{item.schoolName}</h2>
                  <ArrowRight className="w-4 h-4 text-slate-400 group-hover:text-[#008272] transition" />
                </div>
              )}

              {item.subDetails === 'Last used' && (
                <h2 className="font-bold text-slate-900 text-base mb-1">{item.schoolName}</h2>
              )}

              {item.type === 'pending' && (
                <div className="flex items-center justify-between mb-1">
                  <h2 className="font-bold text-slate-900 text-base">{item.schoolName}</h2>
                  <Lock className="w-4 h-4 text-slate-400" />
                </div>
              )}

              {/* Role Badge */}
              <div className="flex items-center gap-2 mb-2.5">
                {item.role === 'Parent' && (
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-700">
                    Parent
                  </span>
                )}
                {item.role === 'Teacher' && (
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-700">
                    Teacher
                  </span>
                )}
                {item.role === 'Waiting for verification' && (
                  <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100/80 text-purple-700">
                    <Clock className="w-3 h-3" /> Waiting for verification
                  </span>
                )}

                {item.role === 'Teacher' && item.details.includes('Class teacher') && (
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-50 text-purple-600">
                    Class teacher • Grade 8 East
                  </span>
                )}
              </div>

              {/* Detail Icons & Text */}
              <div className="flex items-center gap-2 text-xs text-slate-500 pt-1 border-t border-slate-100">
                {item.role === 'Parent' && <Users className="w-3.5 h-3.5 text-slate-400 shrink-0" />}
                {item.role === 'Teacher' && item.details.includes('Subjects: Mathematics') && <BookOpen className="w-3.5 h-3.5 text-slate-400 shrink-0" />}
                {item.role === 'Teacher' && item.details.includes('Subjects: Physics') && <FlaskConical className="w-3.5 h-3.5 text-slate-400 shrink-0" />}
                {item.type === 'pending' && <Clock className="w-3.5 h-3.5 text-slate-400 shrink-0" />}
                
                <span className="truncate">{item.details}</span>
              </div>
            </div>
          ))}
        </div>

        {/* Missing Institutional Profile Footer Link */}
        <div className="pt-2 border-t border-dashed border-slate-300 text-center">
          <button className="inline-flex items-center gap-1.5 text-xs text-slate-500 hover:text-[#008272] transition font-medium">
            <HelpCircle className="w-3.5 h-3.5 text-slate-400" />
            Missing an institutional profile? Contact school registrar
          </button>
        </div>

      </div>

      {/* Bottom Floating Navigation Bar */}
      <div className="w-full max-w-md bg-white border border-slate-200/90 rounded-2xl p-2 shadow-lg flex items-center justify-around text-xs font-medium text-slate-500 mb-2">
        <button 
          onClick={() => setActiveNav('home')} 
          className={`flex flex-col items-center gap-1 py-1 px-4 rounded-xl transition ${activeNav === 'home' ? 'text-[#008272]' : 'hover:text-slate-800'}`}
        >
          <LayoutGrid className="w-5 h-5" />
          <span>Home</span>
        </button>
        <button 
          onClick={() => setActiveNav('roles')} 
          className={`flex flex-col items-center gap-1 py-1 px-4 rounded-xl transition ${activeNav === 'roles' ? 'text-[#008272] font-semibold' : 'hover:text-slate-800'}`}
        >
          <Users className="w-5 h-5" />
          <span>Roles</span>
        </button>
        <button 
          onClick={() => setActiveNav('alerts')} 
          className={`flex flex-col items-center gap-1 py-1 px-4 rounded-xl transition ${activeNav === 'alerts' ? 'text-[#008272]' : 'hover:text-slate-800'}`}
        >
          <Bell className="w-5 h-5" />
          <span>Alerts</span>
        </button>
        <button 
          onClick={() => setActiveNav('settings')} 
          className={`flex flex-col items-center gap-1 py-1 px-4 rounded-xl transition ${activeNav === 'settings' ? 'text-[#008272]' : 'hover:text-slate-800'}`}
        >
          <SettingsIcon className="w-5 h-5" />
          <span>Settings</span>
        </button>
      </div>

    </div>
  );
}